import { expect } from 'chai';
import { spec } from 'modules/tneBidAdapter.js';
import { BANNER, VIDEO } from 'src/mediaTypes.js';

describe('TNE Bid Adapter', function () {
  const ENDPOINT = 'https://catalyst.springwire.ai/openrtb2/auction';

  // ---------------------------------------------------------------------------
  // Fixtures
  // ---------------------------------------------------------------------------

  function makeBannerBidRequest(overrides = {}) {
    return {
      bidder: 'tne',
      params: {
        publisherId: 'pub-123',
        ...overrides.params,
      },
      mediaTypes: {
        banner: {
          sizes: [[300, 250], [728, 90]],
        },
      },
      adUnitCode: 'banner-div',
      bidId: 'bid-banner-1',
      bidderRequestId: 'bidder-req-1',
      auctionId: 'auction-1',
      transactionId: 'txn-1',
      ...overrides,
    };
  }

  function makeVideoBidRequest(overrides = {}) {
    return {
      bidder: 'tne',
      params: {
        publisherId: 'pub-123',
        placementId: 'video-hero',
        ...overrides.params,
      },
      mediaTypes: {
        video: {
          context: 'instream',
          playerSize: [[640, 480]],
          mimes: ['video/mp4'],
          protocols: [2, 3, 5, 6],
          minduration: 5,
          maxduration: 30,
        },
      },
      adUnitCode: 'video-div',
      bidId: 'bid-video-1',
      bidderRequestId: 'bidder-req-1',
      auctionId: 'auction-1',
      transactionId: 'txn-2',
      ...overrides,
    };
  }

  function makeBidderRequest(overrides = {}) {
    return {
      bidderCode: 'tne',
      auctionId: 'auction-1',
      bidderRequestId: 'bidder-req-1',
      timeout: 3000,
      refererInfo: {
        page: 'https://example.com/article',
        domain: 'example.com',
        ref: 'https://google.com',
        topmostLocation: 'https://example.com/article',
      },
      ...overrides,
    };
  }

  function makeServerResponse(overrides = {}) {
    return {
      body: {
        id: 'auction-1',
        seatbid: [
          {
            seat: 'tne',
            bid: [
              {
                id: 'bid-resp-1',
                impid: 'bid-banner-1',
                price: 2.5,
                adm: '<div>ad markup</div>',
                adomain: ['advertiser.com'],
                crid: 'creative-123',
                w: 300,
                h: 250,
                mtype: 1,
                ...overrides.bid,
              },
            ],
          },
        ],
        cur: 'USD',
        ...overrides.response,
      },
    };
  }

  function makeVideoServerResponse(overrides = {}) {
    return {
      body: {
        id: 'auction-1',
        seatbid: [
          {
            seat: 'tne',
            bid: [
              {
                id: 'bid-resp-2',
                impid: 'bid-video-1',
                price: 5.0,
                adm: '<VAST version="4.0"><Ad></Ad></VAST>',
                adomain: ['video-advertiser.com'],
                crid: 'creative-video-456',
                w: 640,
                h: 480,
                mtype: 2,
                ...overrides.bid,
              },
            ],
          },
        ],
        cur: 'USD',
        ...overrides.response,
      },
    };
  }

  // ---------------------------------------------------------------------------
  // isBidRequestValid
  // ---------------------------------------------------------------------------

  describe('isBidRequestValid', function () {
    it('should return true when publisherId is present', function () {
      const bid = makeBannerBidRequest();
      expect(spec.isBidRequestValid(bid)).to.be.true;
    });

    it('should return false when params is missing', function () {
      const bid = makeBannerBidRequest();
      delete bid.params;
      expect(spec.isBidRequestValid(bid)).to.be.false;
    });

    it('should return false when publisherId is missing', function () {
      const bid = makeBannerBidRequest({ params: { publisherId: undefined } });
      delete bid.params.publisherId;
      expect(spec.isBidRequestValid(bid)).to.be.false;
    });

    it('should return false when publisherId is empty string', function () {
      const bid = makeBannerBidRequest({ params: { publisherId: '' } });
      expect(spec.isBidRequestValid(bid)).to.be.false;
    });

    it('should return true for video bid with publisherId', function () {
      const bid = makeVideoBidRequest();
      expect(spec.isBidRequestValid(bid)).to.be.true;
    });

    it('should return true when optional placementId is present', function () {
      const bid = makeBannerBidRequest({
        params: { publisherId: 'pub-456', placementId: 'sidebar' },
      });
      expect(spec.isBidRequestValid(bid)).to.be.true;
    });
  });

  // ---------------------------------------------------------------------------
  // buildRequests
  // ---------------------------------------------------------------------------

  describe('buildRequests', function () {
    it('should return a valid POST request to the TNE endpoint', function () {
      const bidRequests = [makeBannerBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);

      expect(requests).to.be.an('array').that.has.lengthOf(1);
      expect(requests[0].method).to.equal('POST');
      expect(requests[0].url).to.equal(ENDPOINT);
      expect(requests[0].options.contentType).to.equal('application/json');
      expect(requests[0].options.withCredentials).to.be.true;
    });

    it('should build a valid OpenRTB request object', function () {
      const bidRequests = [makeBannerBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const data = requests[0].data;

      expect(data).to.have.property('id');
      expect(data).to.have.property('imp').that.is.an('array');
      expect(data.imp).to.have.lengthOf(1);
    });

    it('should set site.publisher.id from publisherId param', function () {
      const bidRequests = [makeBannerBidRequest({ params: { publisherId: 'pub-789' } })];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const data = requests[0].data;

      expect(data.site.publisher.id).to.equal('pub-789');
    });

    it('should set imp.tagid from placementId param', function () {
      const bidRequests = [
        makeBannerBidRequest({
          params: { publisherId: 'pub-123', placementId: 'sidebar-300x250' },
        }),
      ];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const data = requests[0].data;

      expect(data.imp[0].tagid).to.equal('sidebar-300x250');
    });

    it('should set default currency to USD', function () {
      const bidRequests = [makeBannerBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const data = requests[0].data;

      expect(data.cur).to.deep.equal(['USD']);
    });

    it('should include banner impression data', function () {
      const bidRequests = [makeBannerBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const imp = requests[0].data.imp[0];

      expect(imp).to.have.property('banner');
    });

    it('should include video impression data', function () {
      const bidRequests = [makeVideoBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const imp = requests[0].data.imp[0];

      expect(imp).to.have.property('video');
      expect(imp.video.mimes).to.deep.equal(['video/mp4']);
    });

    it('should handle multiple bid requests in a single ORTB request', function () {
      const bidRequests = [makeBannerBidRequest(), makeVideoBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);

      expect(requests).to.have.lengthOf(1);
      expect(requests[0].data.imp).to.have.lengthOf(2);
    });

    it('should apply bidFloor from params when provided', function () {
      const bidRequests = [
        makeBannerBidRequest({
          params: { publisherId: 'pub-123', bidFloor: 1.5, bidFloorCur: 'USD' },
        }),
      ];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const imp = requests[0].data.imp[0];

      expect(imp.bidfloor).to.equal(1.5);
      expect(imp.bidfloorcur).to.equal('USD');
    });

    it('should pass custom params in imp.ext.bidder.custom', function () {
      const bidRequests = [
        makeBannerBidRequest({
          params: { publisherId: 'pub-123', custom: { section: 'sports' } },
        }),
      ];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const imp = requests[0].data.imp[0];

      expect(imp.ext.bidder.custom).to.deep.equal({ section: 'sports' });
    });

    it('should include adapter version in ext.prebid', function () {
      const bidRequests = [makeBannerBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const data = requests[0].data;

      expect(data.ext.prebid.adapterVersion).to.be.a('string');
    });

    it('should forward GDPR consent when present', function () {
      const bidRequests = [makeBannerBidRequest()];
      const bidderRequest = makeBidderRequest({
        gdprConsent: {
          gdprApplies: true,
          consentString: 'CPXxRfAPXxRfAAfKABENB-CgAAAAAAAAAAYgAAAAAAAA',
        },
      });
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const data = requests[0].data;

      expect(data.regs.ext.gdpr).to.equal(1);
      expect(data.user.ext.consent).to.equal(
        'CPXxRfAPXxRfAAfKABENB-CgAAAAAAAAAAYgAAAAAAAA'
      );
    });

    it('should forward USP consent when present', function () {
      const bidRequests = [makeBannerBidRequest()];
      const bidderRequest = makeBidderRequest({
        uspConsent: '1YNN',
      });
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const data = requests[0].data;

      expect(data.regs.ext.us_privacy).to.equal('1YNN');
    });
  });

  // ---------------------------------------------------------------------------
  // interpretResponse
  // ---------------------------------------------------------------------------

  describe('interpretResponse', function () {
    it('should return an empty array when response body is empty', function () {
      const bids = spec.interpretResponse({ body: null }, { data: {} });
      expect(bids).to.be.an('array').that.is.empty;
    });

    it('should return an empty array when response is undefined', function () {
      const bids = spec.interpretResponse(undefined, { data: {} });
      expect(bids).to.be.an('array').that.is.empty;
    });

    it('should parse a banner bid response correctly', function () {
      const bidRequests = [makeBannerBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const serverResponse = makeServerResponse();

      const bids = spec.interpretResponse(serverResponse, requests[0]);

      expect(bids).to.be.an('array').that.has.lengthOf(1);
      const bid = bids[0];
      expect(bid.cpm).to.equal(2.5);
      expect(bid.width).to.equal(300);
      expect(bid.height).to.equal(250);
      expect(bid.creativeId).to.equal('creative-123');
      expect(bid.currency).to.equal('USD');
      expect(bid.netRevenue).to.be.true;
      expect(bid.ttl).to.equal(300);
      expect(bid.ad).to.equal('<div>ad markup</div>');
      expect(bid.meta.advertiserDomains).to.deep.equal(['advertiser.com']);
    });

    it('should parse a video bid response with inline VAST', function () {
      const bidRequests = [makeVideoBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const serverResponse = makeVideoServerResponse();

      const bids = spec.interpretResponse(serverResponse, requests[0]);

      expect(bids).to.be.an('array').that.has.lengthOf(1);
      const bid = bids[0];
      expect(bid.cpm).to.equal(5.0);
      expect(bid.mediaType).to.equal(VIDEO);
      expect(bid.vastXml).to.contain('<VAST');
    });

    it('should parse a video bid response with nurl when adm is absent', function () {
      const bidRequests = [makeVideoBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);
      const serverResponse = makeVideoServerResponse({
        bid: {
          adm: undefined,
          nurl: 'https://catalyst.springwire.ai/vast?id=abc',
        },
      });

      const bids = spec.interpretResponse(serverResponse, requests[0]);

      expect(bids).to.be.an('array').that.has.lengthOf(1);
      const bid = bids[0];
      expect(bid.vastUrl).to.equal('https://catalyst.springwire.ai/vast?id=abc');
    });

    it('should handle multiple seat bids', function () {
      const bidRequests = [makeBannerBidRequest(), makeVideoBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);

      const serverResponse = {
        body: {
          id: 'auction-1',
          seatbid: [
            {
              seat: 'tne',
              bid: [
                {
                  id: 'bid-resp-1',
                  impid: 'bid-banner-1',
                  price: 2.5,
                  adm: '<div>banner ad</div>',
                  adomain: ['advertiser.com'],
                  crid: 'creative-banner',
                  w: 300,
                  h: 250,
                  mtype: 1,
                },
              ],
            },
            {
              seat: 'tne',
              bid: [
                {
                  id: 'bid-resp-2',
                  impid: 'bid-video-1',
                  price: 5.0,
                  adm: '<VAST version="4.0"></VAST>',
                  adomain: ['video-advertiser.com'],
                  crid: 'creative-video',
                  w: 640,
                  h: 480,
                  mtype: 2,
                },
              ],
            },
          ],
          cur: 'USD',
        },
      };

      const bids = spec.interpretResponse(serverResponse, requests[0]);
      expect(bids).to.have.lengthOf(2);
    });

    it('should return empty array when seatbid is empty', function () {
      const bidRequests = [makeBannerBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);

      const serverResponse = {
        body: {
          id: 'auction-1',
          seatbid: [],
          cur: 'USD',
        },
      };

      const bids = spec.interpretResponse(serverResponse, requests[0]);
      expect(bids).to.be.an('array').that.is.empty;
    });
  });

  // ---------------------------------------------------------------------------
  // getUserSyncs
  // ---------------------------------------------------------------------------

  describe('getUserSyncs', function () {
    const gdprConsent = {
      gdprApplies: true,
      consentString: 'CPXxRfAPXxRfAAfKABENB-CgAAAAAAAAAAYgAAAAAAAA',
    };
    const uspConsent = '1YNN';
    const gppConsent = {
      gppString: 'DBACNYA~CPXxRfAPXxRfAAfKABENB-CgAAAAAAAAAAYgAAAAAAAA',
      applicableSections: [7, 8],
    };

    it('should return an iframe sync when iframeEnabled is true', function () {
      const syncs = spec.getUserSyncs({ iframeEnabled: true }, [], null, null);
      expect(syncs).to.have.lengthOf(1);
      expect(syncs[0].type).to.equal('iframe');
      expect(syncs[0].url).to.contain('catalyst.springwire.ai/usersync/iframe');
    });

    it('should return a pixel sync when pixelEnabled is true', function () {
      const syncs = spec.getUserSyncs({ pixelEnabled: true }, [], null, null);
      expect(syncs).to.have.lengthOf(1);
      expect(syncs[0].type).to.equal('image');
      expect(syncs[0].url).to.contain('catalyst.springwire.ai/setuid');
    });

    it('should include bidder code in pixel sync URL', function () {
      const syncs = spec.getUserSyncs({ pixelEnabled: true }, [], null, null);
      expect(syncs[0].url).to.contain('bidder=tne');
    });

    it('should return both syncs when both are enabled', function () {
      const syncs = spec.getUserSyncs(
        { iframeEnabled: true, pixelEnabled: true },
        [],
        null,
        null
      );
      expect(syncs).to.have.lengthOf(2);
      expect(syncs[0].type).to.equal('iframe');
      expect(syncs[1].type).to.equal('image');
    });

    it('should return empty array when no sync options are enabled', function () {
      const syncs = spec.getUserSyncs(
        { iframeEnabled: false, pixelEnabled: false },
        [],
        null,
        null
      );
      expect(syncs).to.have.lengthOf(0);
    });

    it('should return empty array when COPPA applies', function () {
      const syncs = spec.getUserSyncs(
        { iframeEnabled: true, pixelEnabled: true, coppa: true },
        [],
        null,
        null
      );
      expect(syncs).to.have.lengthOf(0);
    });

    it('should append GDPR consent params to sync URLs', function () {
      const syncs = spec.getUserSyncs(
        { iframeEnabled: true },
        [],
        gdprConsent,
        null
      );
      expect(syncs[0].url).to.contain('gdpr=1');
      expect(syncs[0].url).to.contain('gdpr_consent=');
    });

    it('should append USP consent param to sync URLs', function () {
      const syncs = spec.getUserSyncs(
        { iframeEnabled: true },
        [],
        null,
        uspConsent
      );
      expect(syncs[0].url).to.contain('us_privacy=1YNN');
    });

    it('should append GPP consent params to sync URLs', function () {
      const syncs = spec.getUserSyncs(
        { iframeEnabled: true },
        [],
        null,
        null,
        gppConsent
      );
      expect(syncs[0].url).to.contain('gpp=');
      expect(syncs[0].url).to.contain('gpp_sid=');
    });

    it('should set gdpr=0 when gdprApplies is false', function () {
      const syncs = spec.getUserSyncs(
        { iframeEnabled: true },
        [],
        { gdprApplies: false, consentString: '' },
        null
      );
      expect(syncs[0].url).to.contain('gdpr=0');
    });

    it('should handle all consent params together', function () {
      const syncs = spec.getUserSyncs(
        { iframeEnabled: true },
        [],
        gdprConsent,
        uspConsent,
        gppConsent
      );
      const url = syncs[0].url;
      expect(url).to.contain('gdpr=1');
      expect(url).to.contain('gdpr_consent=');
      expect(url).to.contain('us_privacy=');
      expect(url).to.contain('gpp=');
      expect(url).to.contain('gpp_sid=');
    });

    it('should include consent params in pixel sync URL', function () {
      const syncs = spec.getUserSyncs(
        { pixelEnabled: true },
        [],
        gdprConsent,
        uspConsent
      );
      const url = syncs[0].url;
      expect(url).to.contain('bidder=tne');
      expect(url).to.contain('gdpr=1');
      expect(url).to.contain('us_privacy=');
    });
  });

  // ---------------------------------------------------------------------------
  // Spec properties
  // ---------------------------------------------------------------------------

  describe('spec properties', function () {
    it('should have correct bidder code', function () {
      expect(spec.code).to.equal('tne');
    });

    it('should support banner and video media types', function () {
      expect(spec.supportedMediaTypes).to.deep.equal([BANNER, VIDEO]);
    });

    it('should have GVL ID 1494', function () {
      expect(spec.gvlid).to.equal(1494);
    });

    it('should declare tneCatalyst alias', function () {
      expect(spec.aliases).to.be.an('array');
      const codes = spec.aliases.map((a) => a.code || a);
      expect(codes).to.include('tneCatalyst');
    });
  });

  // ---------------------------------------------------------------------------
  // Alias / custom endpoint
  // ---------------------------------------------------------------------------

  describe('alias with custom endpoint', function () {
    const CUSTOM_HOST = 'https://exchange.customdomain.com';

    it('should use the default endpoint when no endpoint param is provided', function () {
      const bidRequests = [makeBannerBidRequest()];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);

      expect(requests[0].url).to.equal(ENDPOINT);
    });

    it('should use a custom endpoint when provided in params', function () {
      const bidRequests = [
        makeBannerBidRequest({
          params: { publisherId: 'pub-custom', endpoint: CUSTOM_HOST },
        }),
      ];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);

      expect(requests[0].url).to.equal(`${CUSTOM_HOST}/openrtb2/auction`);
    });

    it('should strip trailing slash from custom endpoint', function () {
      const bidRequests = [
        makeBannerBidRequest({
          params: { publisherId: 'pub-custom', endpoint: `${CUSTOM_HOST}/` },
        }),
      ];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);

      expect(requests[0].url).to.equal(`${CUSTOM_HOST}/openrtb2/auction`);
    });

    it('should still set publisher ID on custom endpoint requests', function () {
      const bidRequests = [
        makeBannerBidRequest({
          params: { publisherId: 'pub-custom', endpoint: CUSTOM_HOST },
        }),
      ];
      const bidderRequest = makeBidderRequest();
      const requests = spec.buildRequests(bidRequests, bidderRequest);

      expect(requests[0].data.site.publisher.id).to.equal('pub-custom');
    });
  });
});
