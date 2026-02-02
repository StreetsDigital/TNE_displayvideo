/**
 * Mock Google Publisher Tag (GPT) for unit testing
 */
function createGPTMock() {
  var _slots = {};
  var _services = [];
  var _initialLoadDisabled = false;
  var _servicesEnabled = false;
  var _displayed = [];
  var _refreshed = [];

  var pubads = {
    disableInitialLoad: function() { _initialLoadDisabled = true; },
    refresh: function(slots) {
      _refreshed = _refreshed.concat(slots || []);
    },
    getTargeting: function(key) { return []; },
    _wasInitialLoadDisabled: function() { return _initialLoadDisabled; },
    _getRefreshed: function() { return _refreshed; }
  };

  var googletag = {
    cmd: [],
    apiReady: true,
    defineSlot: function(adUnitPath, sizes, div) {
      var slot = {
        _adUnitPath: adUnitPath,
        _sizes: sizes,
        _div: div,
        _services: [],
        _targeting: {},
        addService: function(service) {
          this._services.push(service);
          return this;
        },
        setTargeting: function(key, value) {
          this._targeting[key] = value;
          return this;
        },
        getTargeting: function(key) {
          return this._targeting[key] || [];
        }
      };
      _slots[div] = slot;
      return slot;
    },
    pubads: function() { return pubads; },
    enableServices: function() { _servicesEnabled = true; },
    display: function(divOrSlot) {
      _displayed.push(divOrSlot);
    },

    // Test helpers
    _getSlots: function() { return _slots; },
    _getSlot: function(code) { return _slots[code]; },
    _wasServicesEnabled: function() { return _servicesEnabled; },
    _getDisplayed: function() { return _displayed; },
    _reset: function() {
      _slots = {};
      _services = [];
      _initialLoadDisabled = false;
      _servicesEnabled = false;
      _displayed = [];
      _refreshed = [];
    },

    // Process queued commands
    _processQueue: function() {
      var q = this.cmd.slice();
      this.cmd = [];
      q.forEach(function(fn) { fn(); });
    }
  };

  return googletag;
}

module.exports = { createGPTMock: createGPTMock };
