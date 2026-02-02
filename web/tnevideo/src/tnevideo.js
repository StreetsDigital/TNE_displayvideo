/**
 * TNEVideo Client SDK v1.0.0
 *
 * A thin orchestration wrapper around Prebid.js and Google Publisher Tag (GPT)
 * that connects publisher pages to the TNE Catalyst Prebid Server.
 *
 * Usage:
 *   <script src="tnevideo.js"></script>
 *   <script>
 *     TNEVideo.init({
 *       serverUrl: 'https://pbs.yourdomain.com',
 *       publisherId: 'pub-12345',
 *       adUnits: [{ code: 'banner-1', mediaTypes: { banner: { sizes: [[300,250]] } }, bids: { appnexus: { placementId: 123 } } }]
 *     });
 *   </script>
 */
(function(window) {
  'use strict';

  // ────────────────────────────────────────────────────────────────
  // Internal state
  // ────────────────────────────────────────────────────────────────

  var _config = null;
  var _initialized = false;
  var _debug = false;
  var _gamSlots = {};
  var _scriptsLoaded = false;

  // ────────────────────────────────────────────────────────────────
  // Debug Logger
  // ────────────────────────────────────────────────────────────────

  function log() {
    if (_debug && window.console && window.console.log) {
      var args = Array.prototype.slice.call(arguments);
      args.unshift('[TNEVideo]');
      window.console.log.apply(window.console, args);
    }
  }

  function warn() {
    if (window.console && window.console.warn) {
      var args = Array.prototype.slice.call(arguments);
      args.unshift('[TNEVideo]');
      window.console.warn.apply(window.console, args);
    }
  }

  function logError() {
    if (window.console && window.console.error) {
      var args = Array.prototype.slice.call(arguments);
      args.unshift('[TNEVideo]');
      window.console.error.apply(window.console, args);
    }
  }

  // ────────────────────────────────────────────────────────────────
  // Utility Functions
  // ────────────────────────────────────────────────────────────────

  function isObject(val) {
    return val !== null && typeof val === 'object' && !Array.isArray(val);
  }

  function deepClone(obj) {
    if (obj === null || typeof obj !== 'object') return obj;
    if (Array.isArray(obj)) return obj.map(deepClone);
    var clone = {};
    for (var key in obj) {
      if (obj.hasOwnProperty(key)) {
        clone[key] = deepClone(obj[key]);
      }
    }
    return clone;
  }

  function mergeObjects(base, override) {
    var result = deepClone(base);
    for (var key in override) {
      if (override.hasOwnProperty(key)) {
        if (isObject(result[key]) && isObject(override[key])) {
          result[key] = mergeObjects(result[key], override[key]);
        } else {
          result[key] = deepClone(override[key]);
        }
      }
    }
    return result;
  }

  // ────────────────────────────────────────────────────────────────
  // Default Configuration
  // ────────────────────────────────────────────────────────────────

  var DEFAULTS = {
    gamNetworkId: '',
    timeout: 1500,
    s2sTimeout: 1000,
    s2sBidders: ['appnexus', 'rubicon', 'pubmatic'],
    enableSendAllBids: false,
    prebidUrl: 'https://cdn.jsdelivr.net/npm/prebid.js@latest/dist/not-for-prod/prebid.js',
    gptUrl: 'https://securepubads.g.doubleclick.net/tag/js/gpt.js',
    userIds: [
      { name: 'sharedId', storage: { name: '_sharedid', type: 'cookie', expires: 365 } },
      { name: 'pubProvidedId' }
    ],
    consentManagement: {
      gdpr: { cmpApi: 'iab', timeout: 3000, defaultGdprScope: true },
      usp: { cmpApi: 'iab', timeout: 1000 }
    },
    video: {},
    debug: false
  };

  // ────────────────────────────────────────────────────────────────
  // Config Validation
  // ────────────────────────────────────────────────────────────────

  function validateConfig(config) {
    if (!config || typeof config !== 'object') {
      return { valid: false, error: 'Config must be an object' };
    }

    if (!config.serverUrl || typeof config.serverUrl !== 'string') {
      return { valid: false, error: 'serverUrl is required' };
    }

    if (!config.publisherId || typeof config.publisherId !== 'string') {
      return { valid: false, error: 'publisherId is required' };
    }

    if (!Array.isArray(config.adUnits) || config.adUnits.length === 0) {
      return { valid: false, error: 'adUnits must be a non-empty array' };
    }

    // Validate each ad unit
    for (var i = 0; i < config.adUnits.length; i++) {
      var unit = config.adUnits[i];
      if (!unit.code || typeof unit.code !== 'string') {
        return { valid: false, error: 'adUnits[' + i + '].code must be a non-empty string' };
      }
      if (!unit.mediaTypes || typeof unit.mediaTypes !== 'object') {
        return { valid: false, error: 'adUnits[' + i + '].mediaTypes is required' };
      }
      if (!unit.mediaTypes.banner && !unit.mediaTypes.video && !unit.mediaTypes.native) {
        return { valid: false, error: 'adUnits[' + i + '].mediaTypes must contain at least one of: banner, video, native' };
      }
      if (!unit.bids || typeof unit.bids !== 'object' || Object.keys(unit.bids).length === 0) {
        return { valid: false, error: 'adUnits[' + i + '].bids must be a non-empty object' };
      }
    }

    // Merge defaults with user config
    var merged = mergeObjects(DEFAULTS, config);

    // Normalize serverUrl: strip trailing slash
    merged.serverUrl = merged.serverUrl.replace(/\/+$/, '');

    // Allow explicit null for consentManagement
    if (config.consentManagement === null) {
      merged.consentManagement = null;
    }

    return { valid: true, config: merged };
  }

  // ────────────────────────────────────────────────────────────────
  // Script Loader
  // ────────────────────────────────────────────────────────────────

  var SCRIPT_LOAD_TIMEOUT = 10000; // 10 seconds

  function loadScript(url, callback) {
    var script = document.createElement('script');
    var done = false;
    var timer = null;

    function finish(err) {
      if (done) return;
      done = true;
      if (timer) clearTimeout(timer);
      callback(err);
    }

    script.type = 'text/javascript';
    script.async = true;
    script.src = url;

    script.onload = function() {
      finish(null);
    };

    script.onerror = function() {
      finish(new Error('Failed to load script: ' + url));
    };

    timer = setTimeout(function() {
      finish(new Error('Script load timeout: ' + url));
    }, SCRIPT_LOAD_TIMEOUT);

    var head = document.head || document.getElementsByTagName('head')[0];
    head.appendChild(script);
  }

  function loadPrebid(url, callback) {
    if (window.pbjs && window.pbjs.setConfig) {
      log('Prebid.js already loaded');
      callback(null);
      return;
    }

    window.pbjs = window.pbjs || {};
    window.pbjs.que = window.pbjs.que || [];

    log('Loading Prebid.js from', url);
    loadScript(url, function(err) {
      if (err) {
        callback(new Error('Failed to load Prebid.js from ' + url));
      } else {
        log('Prebid.js loaded successfully');
        callback(null);
      }
    });
  }

  function loadGPT(url, callback) {
    if (window.googletag && window.googletag.apiReady) {
      log('GPT already loaded');
      callback(null);
      return;
    }

    window.googletag = window.googletag || {};
    window.googletag.cmd = window.googletag.cmd || [];

    log('Loading GPT from', url);
    loadScript(url, function(err) {
      if (err) {
        callback(new Error('Failed to load GPT from ' + url));
      } else {
        log('GPT loaded successfully');
        callback(null);
      }
    });
  }

  function loadScriptsParallel(tasks, finalCallback) {
    var remaining = tasks.length;
    var errors = [];

    if (remaining === 0) {
      finalCallback([]);
      return;
    }

    tasks.forEach(function(task) {
      task.fn(function(err) {
        if (err) {
          errors.push(err);
          logError(task.name + ' load error:', err.message);
        }
        remaining--;
        if (remaining === 0) {
          finalCallback(errors);
        }
      });
    });
  }

  // ────────────────────────────────────────────────────────────────
  // Prebid.js Configuration Builder
  // ────────────────────────────────────────────────────────────────

  function buildPrebidConfig(config) {
    var pbjsConfig = {
      debug: config.debug,
      s2sConfig: {
        accountId: config.publisherId,
        bidders: config.s2sBidders,
        timeout: config.s2sTimeout,
        adapter: 'prebidServer',
        endpoint: {
          p1Consent: config.serverUrl + '/openrtb2/auction',
          noP1Consent: config.serverUrl + '/openrtb2/auction'
        },
        syncEndpoint: {
          p1Consent: config.serverUrl + '/cookie_sync',
          noP1Consent: config.serverUrl + '/cookie_sync'
        }
      },
      userSync: {
        userIds: config.userIds,
        syncDelay: 3000
      },
      priceGranularity: 'medium',
      enableSendAllBids: config.enableSendAllBids,
      bidderTimeout: config.timeout
    };

    // Only set consent management if not explicitly null
    if (config.consentManagement !== null) {
      pbjsConfig.consentManagement = config.consentManagement;
    }

    return pbjsConfig;
  }

  // ────────────────────────────────────────────────────────────────
  // Ad Unit Transformer
  // ────────────────────────────────────────────────────────────────

  function transformAdUnits(adUnits, config) {
    return adUnits.map(function(unit) {
      var pbjsUnit = {
        code: unit.code,
        mediaTypes: deepClone(unit.mediaTypes)
      };

      // Merge default video params if this unit has video
      if (pbjsUnit.mediaTypes.video && config.video) {
        pbjsUnit.mediaTypes.video = mergeObjects(config.video, pbjsUnit.mediaTypes.video);
      }

      // Transform bids from { bidder: params } to [{ bidder, params }]
      pbjsUnit.bids = [];
      var bids = unit.bids || {};
      for (var bidder in bids) {
        if (bids.hasOwnProperty(bidder)) {
          pbjsUnit.bids.push({
            bidder: bidder,
            params: bids[bidder]
          });
        }
      }

      return pbjsUnit;
    });
  }

  // ────────────────────────────────────────────────────────────────
  // GAM Slot Manager
  // ────────────────────────────────────────────────────────────────

  function setupGAMSlots(adUnits, config) {
    window.googletag.cmd.push(function() {
      adUnits.forEach(function(unit) {
        // Skip if slot already defined
        if (_gamSlots[unit.code]) {
          log('GAM slot already defined for', unit.code);
          return;
        }

        var adUnitPath = unit.gamAdUnitPath ||
          (config.gamNetworkId ? config.gamNetworkId + '/' + unit.code : unit.code);

        var slot;
        if (unit.mediaTypes.video && unit.mediaTypes.video.context === 'instream') {
          // Instream video: define slot with 1x1 size (targeting only, no div rendering)
          slot = window.googletag.defineSlot(adUnitPath, [1, 1], unit.code);
        } else {
          // Banner / outstream: associate with div element
          var sizes = [];
          if (unit.mediaTypes.banner) {
            sizes = unit.mediaTypes.banner.sizes || [];
          }
          if (unit.mediaTypes.video && unit.mediaTypes.video.context === 'outstream') {
            sizes = sizes.length ? sizes : [unit.mediaTypes.video.playerSize || [640, 360]];
          }
          slot = window.googletag.defineSlot(adUnitPath, sizes, unit.code);
        }

        if (slot) {
          slot.addService(window.googletag.pubads());
          _gamSlots[unit.code] = slot;
          log('GAM slot defined:', adUnitPath, '->', unit.code);
        }
      });

      window.googletag.pubads().disableInitialLoad();
      window.googletag.enableServices();
      log('GAM services enabled');
    });
  }

  function setTargetingAndDisplay(adUnits) {
    window.googletag.cmd.push(function() {
      // Set targeting from Prebid bid responses
      window.pbjs.que.push(function() {
        window.pbjs.setTargetingForGPTAsync();
        log('Targeting set on GAM slots');
      });

      // Display and refresh non-instream slots
      var slotsToRefresh = [];

      adUnits.forEach(function(unit) {
        if (unit.mediaTypes.video && unit.mediaTypes.video.context === 'instream') {
          return; // Instream handled via buildVideoUrl
        }
        window.googletag.display(unit.code);
        if (_gamSlots[unit.code]) {
          slotsToRefresh.push(_gamSlots[unit.code]);
        }
      });

      if (slotsToRefresh.length > 0) {
        window.googletag.pubads().refresh(slotsToRefresh);
        log('Refreshed', slotsToRefresh.length, 'GAM slots');
      }
    });
  }

  // ────────────────────────────────────────────────────────────────
  // Video URL Builders
  // ────────────────────────────────────────────────────────────────

  function buildVideoUrl(params) {
    if (!window.pbjs || !window.pbjs.adServers || !window.pbjs.adServers.dfp) {
      warn('Prebid.js DFP ad server module not available for buildVideoUrl');
      return null;
    }

    if (!params || !params.adUnit) {
      warn('buildVideoUrl requires params.adUnit');
      return null;
    }

    var videoParams = {
      adUnit: params.adUnit,
      params: {
        iu: params.iu || (params.adUnit && params.adUnit.gamAdUnitPath) || '',
        output: 'vast',
        description_url: params.description_url || window.location.href
      }
    };

    if (params.custParams) {
      videoParams.params.cust_params = params.custParams;
    }

    var url = window.pbjs.adServers.dfp.buildVideoUrl(videoParams);
    log('Built video URL:', url);
    return url || null;
  }

  function buildAdpodVideoUrl(params) {
    if (!window.pbjs || !window.pbjs.adServers || !window.pbjs.adServers.dfp) {
      if (params && params.callback) {
        params.callback(new Error('Prebid.js DFP ad server module not available'), null);
      }
      return;
    }

    window.pbjs.adServers.dfp.buildAdpodVideoUrl({
      codes: params.codes,
      params: {
        iu: params.iu,
        description_url: params.descriptionUrl || window.location.href,
        output: 'vast'
      },
      callback: function(err, vastUrl) {
        log('Built adpod video URL:', vastUrl, err ? 'error:' + err : '');
        if (params.callback) {
          params.callback(err, vastUrl);
        }
      }
    });
  }

  // ────────────────────────────────────────────────────────────────
  // Auction Lifecycle
  // ────────────────────────────────────────────────────────────────

  function runAuction(config, callback) {
    log('Starting auction with', config.adUnits.length, 'ad units');

    var pbjsUnits = transformAdUnits(config.adUnits, config);

    window.pbjs.que.push(function() {
      // Configure Prebid.js
      var pbjsConfig = buildPrebidConfig(config);
      window.pbjs.setConfig(pbjsConfig);
      log('Prebid.js configured');

      // Add ad units
      window.pbjs.addAdUnits(pbjsUnits);
      log('Ad units added:', pbjsUnits.length);

      // Register event handlers
      window.pbjs.onEvent('auctionEnd', function(auctionData) {
        log('Auction ended');

        // Set targeting and display after auction
        setTargetingAndDisplay(config.adUnits);

        if (config.onAuctionEnd) {
          try {
            config.onAuctionEnd(auctionData);
          } catch (e) {
            logError('onAuctionEnd callback error:', e);
          }
        }
      });

      window.pbjs.onEvent('bidWon', function(bid) {
        log('Bid won:', bid.bidderCode, bid.cpm);
        if (config.onBidWon) {
          try {
            config.onBidWon(bid);
          } catch (e) {
            logError('onBidWon callback error:', e);
          }
        }
      });

      // Request bids
      window.pbjs.requestBids({
        timeout: config.timeout,
        bidsBackHandler: function(bids) {
          log('Bids returned for', Object.keys(bids).length, 'ad units');
          if (callback) callback(null, bids);
        }
      });
    });
  }

  // ────────────────────────────────────────────────────────────────
  // Public API
  // ────────────────────────────────────────────────────────────────

  var TNEVideo = {
    /**
     * Initialize the integration. Loads Prebid.js and GPT, configures
     * everything, and runs the first auction automatically.
     * @param {Object} config - Publisher configuration
     */
    init: function(config) {
      if (_initialized) {
        warn('Already initialized, ignoring duplicate init() call');
        return;
      }

      var result = validateConfig(config);
      if (!result.valid) {
        logError(result.error);
        if (config && config.onError) {
          try { config.onError(new Error(result.error)); } catch (e) { /* noop */ }
        }
        return;
      }

      _config = result.config;
      _debug = _config.debug;
      _initialized = true;

      log('Initializing with publisherId:', _config.publisherId, 'serverUrl:', _config.serverUrl);
      log('Ad units:', _config.adUnits.length, 'S2S bidders:', _config.s2sBidders.join(', '));

      // Load Prebid.js and GPT in parallel, then run the auction
      loadScriptsParallel([
        { name: 'prebid', fn: function(cb) { loadPrebid(_config.prebidUrl, cb); } },
        { name: 'gpt', fn: function(cb) { loadGPT(_config.gptUrl, cb); } }
      ], function(errors) {
        if (errors.length > 0) {
          errors.forEach(function(e) { logError(e.message); });
          if (_config.onError) {
            try { _config.onError(errors[0]); } catch (e) { /* noop */ }
          }
          // Still try to setup GAM slots for direct fill even if Prebid fails
          if (!errors.some(function(e) { return e.message.indexOf('GPT') !== -1; })) {
            setupGAMSlots(_config.adUnits, _config);
          }
          return;
        }

        _scriptsLoaded = true;
        log('All scripts loaded successfully');

        // Setup GAM slots
        setupGAMSlots(_config.adUnits, _config);

        // Run the first auction
        runAuction(_config, function(err) {
          if (err && _config.onError) {
            try { _config.onError(err); } catch (e) { /* noop */ }
          }
        });
      });
    },

    /**
     * Trigger a new auction. Useful for infinite scroll or single-page apps.
     * @param {Object} [options] - Optional overrides (e.g. { timeout: 2000 })
     */
    refresh: function(options) {
      if (!_initialized) {
        logError('Not initialized. Call TNEVideo.init() first.');
        return;
      }
      if (!_scriptsLoaded) {
        logError('Scripts not yet loaded. Cannot refresh.');
        return;
      }

      var refreshConfig = options ? mergeObjects(_config, options) : _config;

      log('Refreshing auction');
      window.pbjs.que.push(function() {
        window.pbjs.removeAdUnit();
        runAuction(refreshConfig, function(err) {
          if (err && refreshConfig.onError) {
            try { refreshConfig.onError(err); } catch (e) { /* noop */ }
          }
        });
      });
    },

    /**
     * Add ad units after initialization.
     * @param {Object|Array} newUnits - Ad unit(s) to add
     */
    addAdUnits: function(newUnits) {
      if (!_initialized) {
        logError('Not initialized. Call TNEVideo.init() first.');
        return;
      }

      if (!Array.isArray(newUnits)) {
        newUnits = [newUnits];
      }

      _config.adUnits = _config.adUnits.concat(newUnits);

      if (_scriptsLoaded) {
        var pbjsUnits = transformAdUnits(newUnits, _config);
        window.pbjs.que.push(function() {
          window.pbjs.addAdUnits(pbjsUnits);
        });
        setupGAMSlots(newUnits, _config);
      }

      log('Added', newUnits.length, 'ad units');
    },

    /**
     * Build a GAM VAST tag URL for instream video.
     * Returns null if Prebid.js is not ready.
     * @param {Object} params - { adUnit, iu, custParams, description_url }
     * @returns {string|null}
     */
    buildVideoUrl: function(params) {
      if (!_initialized || !_scriptsLoaded) {
        warn('SDK not ready for buildVideoUrl');
        return null;
      }
      return buildVideoUrl(params);
    },

    /**
     * Build a GAM VAST tag URL for ad pods (long-form video).
     * @param {Object} params - { iu, codes, descriptionUrl, callback }
     */
    buildAdpodVideoUrl: function(params) {
      if (!_initialized || !_scriptsLoaded) {
        if (params && params.callback) {
          params.callback(new Error('SDK not ready'), null);
        }
        return;
      }
      buildAdpodVideoUrl(params);
    }
  };

  // Expose to global scope
  window.TNEVideo = TNEVideo;

})(window);
