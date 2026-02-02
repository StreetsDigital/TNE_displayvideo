/**
 * Mock Prebid.js (pbjs) for unit testing
 */
function createPrebidMock() {
  var _config = null;
  var _adUnits = [];
  var _events = {};
  var _targetingSet = false;
  var _requestBidsCalled = false;
  var _requestBidsTimeout = null;

  return {
    que: [],
    setConfig: function(config) {
      _config = config;
    },
    getConfig: function() {
      return _config;
    },
    addAdUnits: function(units) {
      if (Array.isArray(units)) {
        _adUnits = _adUnits.concat(units);
      } else {
        _adUnits.push(units);
      }
    },
    removeAdUnit: function() {
      _adUnits = [];
    },
    getAdUnits: function() {
      return _adUnits;
    },
    onEvent: function(event, handler) {
      _events[event] = _events[event] || [];
      _events[event].push(handler);
    },
    offEvent: function(event, handler) {
      if (_events[event]) {
        _events[event] = _events[event].filter(function(h) { return h !== handler; });
      }
    },
    requestBids: function(opts) {
      _requestBidsCalled = true;
      _requestBidsTimeout = opts.timeout;
      // Simulate async bid response
      var mockBids = {};
      _adUnits.forEach(function(unit) {
        mockBids[unit.code] = {
          bids: [{
            bidderCode: 'appnexus',
            cpm: 5.00,
            width: 300,
            height: 250,
            adserverTargeting: {
              hb_pb: '5.00',
              hb_bidder: 'appnexus',
              hb_format: 'banner',
              hb_size: '300x250'
            }
          }]
        };
      });

      if (opts.bidsBackHandler) {
        setTimeout(function() { opts.bidsBackHandler(mockBids); }, 0);
      }

      // Fire auctionEnd event
      if (_events.auctionEnd) {
        setTimeout(function() {
          _events.auctionEnd.forEach(function(handler) {
            handler({ auctionId: 'mock-auction-1', bidsReceived: [] });
          });
        }, 0);
      }
    },
    setTargetingForGPTAsync: function() {
      _targetingSet = true;
    },
    adServers: {
      dfp: {
        buildVideoUrl: function(params) {
          return 'https://securepubads.g.doubleclick.net/gampad/ads?iu=' +
            (params.params.iu || '') + '&output=vast&hb_uuid=mock-uuid';
        },
        buildAdpodVideoUrl: function(params) {
          if (params.callback) {
            var url = 'https://securepubads.g.doubleclick.net/gampad/ads?iu=' +
              (params.params.iu || '') + '&output=vast';
            setTimeout(function() { params.callback(null, url); }, 0);
          }
        }
      }
    },

    // Test helpers
    _getConfig: function() { return _config; },
    _getAdUnits: function() { return _adUnits; },
    _wasRequestBidsCalled: function() { return _requestBidsCalled; },
    _getRequestBidsTimeout: function() { return _requestBidsTimeout; },
    _wasTargetingSet: function() { return _targetingSet; },
    _reset: function() {
      _config = null;
      _adUnits = [];
      _events = {};
      _targetingSet = false;
      _requestBidsCalled = false;
      _requestBidsTimeout = null;
    },

    // Process queued commands
    _processQueue: function() {
      var q = this.que.slice();
      this.que = [];
      q.forEach(function(fn) { fn(); });
    }
  };
}

module.exports = { createPrebidMock: createPrebidMock };
