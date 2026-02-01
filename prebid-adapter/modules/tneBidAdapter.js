import { ortbConverter } from '../libraries/ortbConverter/converter.js';
import { registerBidder } from '../src/adapters/bidderFactory.js';
import { BANNER, VIDEO } from '../src/mediaTypes.js';
import { deepSetValue, deepAccess, logWarn } from '../src/utils.js';

/**
 * @typedef {import('../src/adapters/bidderFactory.js').BidRequest} BidRequest
 * @typedef {import('../src/adapters/bidderFactory.js').ServerResponse} ServerResponse
 */

const BIDDER_CODE = 'tne';
const ENDPOINT_URL = 'https://catalyst.springwire.ai/openrtb2/auction';
const COOKIE_SYNC_URL = 'https://catalyst.springwire.ai/cookie_sync';
const USERSYNC_URL = 'https://catalyst.springwire.ai/setuid';
const DEFAULT_CURRENCY = 'USD';
const DEFAULT_TTL = 300;
const ADAPTER_VERSION = '1.0.0';

const converter = ortbConverter({
  context: {
    netRevenue: true,
    ttl: DEFAULT_TTL,
  },

  imp(buildImp, bidRequest, context) {
    const imp = buildImp(bidRequest, context);

    // Map placementId to imp.tagid for granular reporting
    const placementId = deepAccess(bidRequest, 'params.placementId');
    if (placementId) {
      imp.tagid = String(placementId);
    }

    // Ensure bidfloor is set if publisher provided one
    const bidFloor = deepAccess(bidRequest, 'params.bidFloor');
    if (bidFloor && !imp.bidfloor) {
      imp.bidfloor = bidFloor;
      imp.bidfloorcur = deepAccess(bidRequest, 'params.bidFloorCur') || DEFAULT_CURRENCY;
    }

    // Pass any custom bidder params into imp.ext.bidder
    const customParams = deepAccess(bidRequest, 'params.custom');
    if (customParams) {
      deepSetValue(imp, 'ext.bidder.custom', customParams);
    }

    return imp;
  },

  request(buildRequest, imps, bidderRequest, context) {
    const request = buildRequest(imps, bidderRequest, context);

    // Set publisher ID from bid params (required by TNE exchange)
    const publisherId = deepAccess(context, 'bidRequests.0.params.publisherId');
    if (publisherId) {
      deepSetValue(request, 'site.publisher.id', String(publisherId));
    }

    // Set currency preference
    if (!request.cur) {
      request.cur = [DEFAULT_CURRENCY];
    }

    // Tag the request source for analytics
    deepSetValue(request, 'ext.prebid.channel', {
      name: 'pbjs',
      version: '$prebid.version$',
    });
    deepSetValue(request, 'ext.prebid.adapterVersion', ADAPTER_VERSION);

    return request;
  },

  bidResponse(buildBidResponse, bid, context) {
    const bidResponse = buildBidResponse(bid, context);

    // Determine media type from the ORTB response mtype field or bid ext
    if (!bidResponse.mediaType) {
      const mtype = deepAccess(bid, 'mtype');
      if (mtype === 2) {
        bidResponse.mediaType = VIDEO;
      } else {
        bidResponse.mediaType = BANNER;
      }
    }

    // Handle video responses — prefer adm (inline VAST) over nurl
    if (bidResponse.mediaType === VIDEO) {
      if (bid.adm) {
        bidResponse.vastXml = bid.adm;
      } else if (bid.nurl) {
        bidResponse.vastUrl = bid.nurl;
      }
    }

    // Ensure advertiser domains are populated for meta
    if (bid.adomain && bid.adomain.length > 0) {
      bidResponse.meta = bidResponse.meta || {};
      bidResponse.meta.advertiserDomains = bid.adomain;
    }

    return bidResponse;
  },
});

export const spec = {
  code: BIDDER_CODE,
  supportedMediaTypes: [BANNER, VIDEO],

  /**
   * Validate bid request params.
   * publisherId is required by the TNE exchange for publisher authentication.
   *
   * @param {BidRequest} bid
   * @returns {boolean}
   */
  isBidRequestValid(bid) {
    if (!bid.params) {
      return false;
    }
    if (!bid.params.publisherId) {
      logWarn(`${BIDDER_CODE}: publisherId is required in bid params`);
      return false;
    }
    return true;
  },

  /**
   * Build OpenRTB 2.5 request(s) to the TNE Catalyst exchange.
   *
   * @param {BidRequest[]} bidRequests
   * @param {*} bidderRequest
   * @returns {Array<{method: string, url: string, data: object}>}
   */
  buildRequests(bidRequests, bidderRequest) {
    const data = converter.toORTB({ bidRequests, bidderRequest });

    return [
      {
        method: 'POST',
        url: ENDPOINT_URL,
        data,
        options: {
          contentType: 'application/json',
          withCredentials: true,
        },
      },
    ];
  },

  /**
   * Parse OpenRTB response from the exchange into Prebid bid objects.
   *
   * @param {ServerResponse} serverResponse
   * @param {object} request - the original request object from buildRequests
   * @returns {Array} Array of bid response objects
   */
  interpretResponse(serverResponse, request) {
    if (!serverResponse || !serverResponse.body) {
      return [];
    }

    const response = serverResponse.body;
    const ortbRequest = request.data;

    const bids = converter.fromORTB({ response, request: ortbRequest }).bids;
    return bids;
  },

  /**
   * Register user sync pixels.
   * Supports both iframe and image syncs with GDPR/USP consent forwarding.
   *
   * @param {object} syncOptions
   * @param {Array} serverResponses
   * @param {object} gdprConsent
   * @param {string} uspConsent
   * @param {object} gppConsent
   * @returns {Array<{type: string, url: string}>}
   */
  getUserSyncs(syncOptions, serverResponses, gdprConsent, uspConsent, gppConsent) {
    const syncs = [];
    const params = [];

    // Append GDPR consent parameters
    if (gdprConsent) {
      params.push(`gdpr=${gdprConsent.gdprApplies ? 1 : 0}`);
      if (gdprConsent.consentString) {
        params.push(`gdpr_consent=${encodeURIComponent(gdprConsent.consentString)}`);
      }
    }

    // Append USP/CCPA consent
    if (uspConsent) {
      params.push(`us_privacy=${encodeURIComponent(uspConsent)}`);
    }

    // Append GPP consent
    if (gppConsent) {
      if (gppConsent.gppString) {
        params.push(`gpp=${encodeURIComponent(gppConsent.gppString)}`);
      }
      if (gppConsent.applicableSections) {
        params.push(`gpp_sid=${encodeURIComponent(gppConsent.applicableSections.join(','))}`);
      }
    }

    const queryString = params.length > 0 ? `?${params.join('&')}` : '';

    if (syncOptions.iframeEnabled) {
      syncs.push({
        type: 'iframe',
        url: `${COOKIE_SYNC_URL}${queryString}`,
      });
    }

    if (syncOptions.pixelEnabled) {
      syncs.push({
        type: 'image',
        url: `${USERSYNC_URL}${queryString}`,
      });
    }

    return syncs;
  },
};

registerBidder(spec);
