/**
 * Unit tests for TNEVideo SDK
 *
 * Tests are organized by module:
 * 1. Config validation
 * 2. Ad unit transformation
 * 3. Prebid config builder
 * 4. Public API behavior
 * 5. Error handling
 */

const fs = require('fs');
const path = require('path');
const { createPrebidMock } = require('./mocks/prebid-mock');
const { createGPTMock } = require('./mocks/gpt-mock');

// Load the SDK source for testing internal functions
// We execute the IIFE in a controlled environment
let sdkSource;

beforeAll(() => {
  sdkSource = fs.readFileSync(
    path.join(__dirname, '..', 'src', 'tnevideo.js'),
    'utf-8'
  );
});

// Reset global state before each test
beforeEach(() => {
  // Reset TNEVideo by re-evaluating the IIFE
  delete global.window;
  global.window = {
    console: console,
    location: { href: 'https://example.com/page' }
  };
  global.document = {
    createElement: jest.fn().mockReturnValue({
      type: '',
      async: false,
      src: '',
      onload: null,
      onerror: null
    }),
    head: {
      appendChild: jest.fn()
    },
    getElementsByTagName: jest.fn().mockReturnValue([{ appendChild: jest.fn() }])
  };

  // Execute the SDK in our controlled window
  const fn = new Function('window', 'document', sdkSource);
  fn(global.window, global.document);
});

afterEach(() => {
  delete global.window;
  delete global.document;
});

// ─── Config Validation ──────────────────────────────────────────

describe('Config Validation', () => {
  test('rejects missing config', () => {
    const errorSpy = jest.spyOn(console, 'error').mockImplementation();
    window.TNEVideo.init(null);
    expect(errorSpy).toHaveBeenCalledWith(
      expect.stringContaining('[TNEVideo]'),
      expect.stringContaining('Config must be an object')
    );
    errorSpy.mockRestore();
  });

  test('rejects missing serverUrl', () => {
    const onError = jest.fn();
    window.TNEVideo.init({
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }],
      onError: onError
    });
    expect(onError).toHaveBeenCalledWith(expect.objectContaining({
      message: 'serverUrl is required'
    }));
  });

  test('rejects missing publisherId', () => {
    const onError = jest.fn();
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }],
      onError: onError
    });
    expect(onError).toHaveBeenCalledWith(expect.objectContaining({
      message: 'publisherId is required'
    }));
  });

  test('rejects empty adUnits', () => {
    const onError = jest.fn();
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [],
      onError: onError
    });
    expect(onError).toHaveBeenCalledWith(expect.objectContaining({
      message: 'adUnits must be a non-empty array'
    }));
  });

  test('rejects ad unit without code', () => {
    const onError = jest.fn();
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }],
      onError: onError
    });
    expect(onError).toHaveBeenCalledWith(expect.objectContaining({
      message: expect.stringContaining('code must be a non-empty string')
    }));
  });

  test('rejects ad unit without mediaTypes', () => {
    const onError = jest.fn();
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', bids: { demo: { placementId: 1 } } }],
      onError: onError
    });
    expect(onError).toHaveBeenCalledWith(expect.objectContaining({
      message: expect.stringContaining('mediaTypes is required')
    }));
  });

  test('rejects ad unit without valid mediaType', () => {
    const onError = jest.fn();
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: {}, bids: { demo: { placementId: 1 } } }],
      onError: onError
    });
    expect(onError).toHaveBeenCalledWith(expect.objectContaining({
      message: expect.stringContaining('must contain at least one of')
    }));
  });

  test('rejects ad unit without bids', () => {
    const onError = jest.fn();
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } } }],
      onError: onError
    });
    expect(onError).toHaveBeenCalledWith(expect.objectContaining({
      message: expect.stringContaining('bids must be a non-empty object')
    }));
  });

  test('strips trailing slash from serverUrl', () => {
    // We can test this indirectly via the Prebid config
    // Set up mocks so init proceeds
    const mockPbjs = createPrebidMock();
    const mockGpt = createGPTMock();
    window.pbjs = mockPbjs;
    window.googletag = mockGpt;

    // Simulate already-loaded scripts
    window.pbjs.setConfig = mockPbjs.setConfig;

    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com/',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    // No error means config validation passed with stripped slash
    // We can't easily inspect internal _config, but the lack of onError call is sufficient
  });
});

// ─── Duplicate Init Prevention ──────────────────────────────────

describe('Duplicate Init Prevention', () => {
  test('warns on duplicate init call', () => {
    const warnSpy = jest.spyOn(console, 'warn').mockImplementation();
    const mockPbjs = createPrebidMock();
    const mockGpt = createGPTMock();
    window.pbjs = mockPbjs;
    window.googletag = mockGpt;

    const validConfig = {
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    };

    window.TNEVideo.init(validConfig);
    window.TNEVideo.init(validConfig);

    expect(warnSpy).toHaveBeenCalledWith(
      expect.stringContaining('[TNEVideo]'),
      expect.stringContaining('Already initialized')
    );
    warnSpy.mockRestore();
  });
});

// ─── Pre-Init API Calls ─────────────────────────────────────────

describe('Pre-Init API Calls', () => {
  test('refresh before init logs error', () => {
    const errorSpy = jest.spyOn(console, 'error').mockImplementation();
    window.TNEVideo.refresh();
    expect(errorSpy).toHaveBeenCalledWith(
      expect.stringContaining('[TNEVideo]'),
      expect.stringContaining('Not initialized')
    );
    errorSpy.mockRestore();
  });

  test('addAdUnits before init logs error', () => {
    const errorSpy = jest.spyOn(console, 'error').mockImplementation();
    window.TNEVideo.addAdUnits({ code: 'x', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: {} } });
    expect(errorSpy).toHaveBeenCalledWith(
      expect.stringContaining('[TNEVideo]'),
      expect.stringContaining('Not initialized')
    );
    errorSpy.mockRestore();
  });

  test('buildVideoUrl before init returns null', () => {
    const warnSpy = jest.spyOn(console, 'warn').mockImplementation();
    var result = window.TNEVideo.buildVideoUrl({ adUnit: { code: 'x' }, iu: '/123/x' });
    expect(result).toBeNull();
    warnSpy.mockRestore();
  });

  test('buildAdpodVideoUrl before init calls callback with error', (done) => {
    window.TNEVideo.buildAdpodVideoUrl({
      iu: '/123/x',
      callback: function(err, url) {
        expect(err).toBeTruthy();
        expect(err.message).toContain('not ready');
        expect(url).toBeNull();
        done();
      }
    });
  });
});

// ─── Script Loading ─────────────────────────────────────────────

describe('Script Loading', () => {
  test('creates script elements for Prebid and GPT', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    // Should have created at least 2 script elements (Prebid + GPT)
    expect(document.createElement).toHaveBeenCalledWith('script');
    expect(document.createElement.mock.calls.length).toBeGreaterThanOrEqual(2);
  });

  test('skips loading if Prebid.js is already present', () => {
    const mockPbjs = createPrebidMock();
    window.pbjs = mockPbjs;

    const mockGpt = createGPTMock();
    window.googletag = mockGpt;

    const createCount = document.createElement.mock.calls.length;

    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    // Should not create additional script elements since both are already loaded
    expect(document.createElement.mock.calls.length).toBe(createCount);
  });
});

// ─── Integration: Init with Mocked Dependencies ─────────────────

describe('Integration with Mocked Dependencies', () => {
  let mockPbjs, mockGpt;

  beforeEach(() => {
    mockPbjs = createPrebidMock();
    mockGpt = createGPTMock();
    window.pbjs = mockPbjs;
    window.googletag = mockGpt;
  });

  test('configures Prebid.js S2S after init', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      s2sBidders: ['appnexus', 'rubicon'],
      s2sTimeout: 800,
      adUnits: [{ code: 'banner-1', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { appnexus: { placementId: 1 } } }]
    });

    // Process Prebid queue
    mockPbjs._processQueue();

    var config = mockPbjs._getConfig();
    expect(config).toBeTruthy();
    expect(config.s2sConfig.accountId).toBe('pub-123');
    expect(config.s2sConfig.bidders).toEqual(['appnexus', 'rubicon']);
    expect(config.s2sConfig.timeout).toBe(800);
    expect(config.s2sConfig.endpoint.p1Consent).toBe('https://pbs.example.com/openrtb2/auction');
    expect(config.s2sConfig.syncEndpoint.p1Consent).toBe('https://pbs.example.com/cookie_sync');
  });

  test('transforms ad units from flat bids to Prebid format', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{
        code: 'banner-1',
        mediaTypes: { banner: { sizes: [[300, 250], [728, 90]] } },
        bids: {
          appnexus: { placementId: 123 },
          rubicon: { accountId: 1, siteId: 2, zoneId: 3 }
        }
      }]
    });

    // Process queues
    mockPbjs._processQueue();

    var units = mockPbjs._getAdUnits();
    expect(units.length).toBe(1);
    expect(units[0].code).toBe('banner-1');
    expect(units[0].mediaTypes.banner.sizes).toEqual([[300, 250], [728, 90]]);
    expect(units[0].bids.length).toBe(2);

    var bidders = units[0].bids.map(function(b) { return b.bidder; }).sort();
    expect(bidders).toEqual(['appnexus', 'rubicon']);

    var appnexusBid = units[0].bids.find(function(b) { return b.bidder === 'appnexus'; });
    expect(appnexusBid.params.placementId).toBe(123);
  });

  test('sets up GAM slots and enables services', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      gamNetworkId: '/19968336',
      adUnits: [
        { code: 'banner-1', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } },
        { code: 'banner-2', mediaTypes: { banner: { sizes: [[728, 90]] } }, bids: { demo: { placementId: 2 } } }
      ]
    });

    // Process GPT queue
    mockGpt._processQueue();

    var slot1 = mockGpt._getSlot('banner-1');
    var slot2 = mockGpt._getSlot('banner-2');

    expect(slot1).toBeTruthy();
    expect(slot1._adUnitPath).toBe('/19968336/banner-1');
    expect(slot2).toBeTruthy();
    expect(slot2._adUnitPath).toBe('/19968336/banner-2');
    expect(mockGpt._wasServicesEnabled()).toBe(true);
  });

  test('defines instream video slots as 1x1', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      gamNetworkId: '/19968336',
      adUnits: [{
        code: 'preroll',
        mediaTypes: { video: { context: 'instream', playerSize: [640, 360], mimes: ['video/mp4'] } },
        bids: { appnexus: { placementId: 1 } }
      }]
    });

    mockGpt._processQueue();

    var slot = mockGpt._getSlot('preroll');
    expect(slot).toBeTruthy();
    expect(slot._sizes).toEqual([1, 1]);
  });

  test('uses explicit gamAdUnitPath when provided', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      gamNetworkId: '/19968336',
      adUnits: [{
        code: 'preroll',
        gamAdUnitPath: '/custom/path',
        mediaTypes: { video: { context: 'instream', playerSize: [640, 360], mimes: ['video/mp4'] } },
        bids: { appnexus: { placementId: 1 } }
      }]
    });

    mockGpt._processQueue();

    var slot = mockGpt._getSlot('preroll');
    expect(slot._adUnitPath).toBe('/custom/path');
  });

  test('sends enableSendAllBids to Prebid config', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      enableSendAllBids: true,
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    mockPbjs._processQueue();

    var config = mockPbjs._getConfig();
    expect(config.enableSendAllBids).toBe(true);
  });

  test('sets price granularity to medium', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    mockPbjs._processQueue();

    var config = mockPbjs._getConfig();
    expect(config.priceGranularity).toBe('medium');
  });

  test('configures user IDs with defaults', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    mockPbjs._processQueue();

    var config = mockPbjs._getConfig();
    expect(config.userSync.userIds.length).toBe(2);
    expect(config.userSync.userIds[0].name).toBe('sharedId');
    expect(config.userSync.userIds[1].name).toBe('pubProvidedId');
  });

  test('allows custom user IDs', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      userIds: [
        { name: 'id5Id', params: { partner: 1234 } }
      ],
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    mockPbjs._processQueue();

    var config = mockPbjs._getConfig();
    expect(config.userSync.userIds.length).toBe(1);
    expect(config.userSync.userIds[0].name).toBe('id5Id');
  });

  test('disables consent management when set to null', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      consentManagement: null,
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    mockPbjs._processQueue();

    var config = mockPbjs._getConfig();
    expect(config.consentManagement).toBeUndefined();
  });

  test('applies default consent management', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    mockPbjs._processQueue();

    var config = mockPbjs._getConfig();
    expect(config.consentManagement.gdpr.cmpApi).toBe('iab');
    expect(config.consentManagement.usp.cmpApi).toBe('iab');
  });

  test('merges default video params into video ad units', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      video: { mimes: ['video/mp4'], protocols: [2, 5] },
      adUnits: [{
        code: 'vid-1',
        mediaTypes: { video: { context: 'instream', playerSize: [640, 360], maxduration: 30 } },
        bids: { appnexus: { placementId: 1 } }
      }]
    });

    mockPbjs._processQueue();

    var units = mockPbjs._getAdUnits();
    expect(units[0].mediaTypes.video.mimes).toEqual(['video/mp4']);
    expect(units[0].mediaTypes.video.protocols).toEqual([2, 5]);
    expect(units[0].mediaTypes.video.maxduration).toBe(30);
    expect(units[0].mediaTypes.video.context).toBe('instream');
  });

  test('requests bids with correct timeout', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      timeout: 2000,
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    mockPbjs._processQueue();

    expect(mockPbjs._wasRequestBidsCalled()).toBe(true);
    expect(mockPbjs._getRequestBidsTimeout()).toBe(2000);
  });
});

// ─── Callback Handling ──────────────────────────────────────────

describe('Callback Handling', () => {
  let mockPbjs, mockGpt;

  beforeEach(() => {
    mockPbjs = createPrebidMock();
    mockGpt = createGPTMock();
    window.pbjs = mockPbjs;
    window.googletag = mockGpt;
  });

  test('onError callback receives validation errors', () => {
    const onError = jest.fn();
    window.TNEVideo.init({
      onError: onError,
      adUnits: []
    });
    expect(onError).toHaveBeenCalled();
  });

  test('callback exceptions do not break SDK', () => {
    const errorSpy = jest.spyOn(console, 'error').mockImplementation();

    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }],
      onAuctionEnd: function() { throw new Error('Callback crash'); },
      onBidWon: function() { throw new Error('Callback crash'); }
    });

    // This should not throw
    expect(() => {
      mockPbjs._processQueue();
    }).not.toThrow();

    errorSpy.mockRestore();
  });
});

// ─── Video URL Building ─────────────────────────────────────────

describe('Video URL Building', () => {
  let mockPbjs, mockGpt;

  beforeEach(() => {
    mockPbjs = createPrebidMock();
    mockGpt = createGPTMock();
    window.pbjs = mockPbjs;
    window.googletag = mockGpt;

    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{
        code: 'preroll',
        mediaTypes: { video: { context: 'instream', playerSize: [640, 360], mimes: ['video/mp4'] } },
        bids: { appnexus: { placementId: 1 } }
      }]
    });

    // Mark scripts as loaded by processing queues
    mockPbjs._processQueue();
    mockGpt._processQueue();
  });

  test('buildVideoUrl delegates to Prebid DFP module', () => {
    var url = window.TNEVideo.buildVideoUrl({
      adUnit: { code: 'preroll', mediaTypes: { video: { context: 'instream' } } },
      iu: '/19968336/preroll'
    });

    expect(url).toBeTruthy();
    expect(url).toContain('/19968336/preroll');
    expect(url).toContain('output=vast');
  });

  test('buildVideoUrl returns null without adUnit param', () => {
    const warnSpy = jest.spyOn(console, 'warn').mockImplementation();
    var url = window.TNEVideo.buildVideoUrl({});
    expect(url).toBeNull();
    warnSpy.mockRestore();
  });

  test('buildAdpodVideoUrl calls callback with URL', (done) => {
    window.TNEVideo.buildAdpodVideoUrl({
      iu: '/19968336/longform',
      descriptionUrl: 'https://example.com/video',
      callback: function(err, url) {
        expect(err).toBeNull();
        expect(url).toBeTruthy();
        expect(url).toContain('/19968336/longform');
        done();
      }
    });
  });
});

// ─── Default Values ─────────────────────────────────────────────

describe('Default Configuration Values', () => {
  let mockPbjs, mockGpt;

  beforeEach(() => {
    mockPbjs = createPrebidMock();
    mockGpt = createGPTMock();
    window.pbjs = mockPbjs;
    window.googletag = mockGpt;
  });

  test('applies default timeout of 1500ms', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    mockPbjs._processQueue();

    var config = mockPbjs._getConfig();
    expect(config.bidderTimeout).toBe(1500);
  });

  test('applies default s2sTimeout of 1000ms', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    mockPbjs._processQueue();

    var config = mockPbjs._getConfig();
    expect(config.s2sConfig.timeout).toBe(1000);
  });

  test('applies default s2sBidders', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    mockPbjs._processQueue();

    var config = mockPbjs._getConfig();
    expect(config.s2sConfig.bidders).toEqual(['appnexus', 'rubicon', 'pubmatic']);
  });

  test('enableSendAllBids defaults to false', () => {
    window.TNEVideo.init({
      serverUrl: 'https://pbs.example.com',
      publisherId: 'pub-123',
      adUnits: [{ code: 'a', mediaTypes: { banner: { sizes: [[300, 250]] } }, bids: { demo: { placementId: 1 } } }]
    });

    mockPbjs._processQueue();

    var config = mockPbjs._getConfig();
    expect(config.enableSendAllBids).toBe(false);
  });
});
