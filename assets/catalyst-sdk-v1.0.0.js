/**
 * Catalyst Bidder SDK for MAI Publisher Integration
 * Server-side header bidding adapter
 * @version 1.0.0
 */

(function(window) {
  'use strict';

  // Catalyst namespace
  var catalyst = window.catalyst || {};
  window.catalyst = catalyst;

  // Command queue for async loading support
  catalyst.cmd = catalyst.cmd || [];

  // Configuration
  catalyst._config = {
    accountId: '',
    serverUrl: window.location.protocol + '//' + window.location.host,
    timeout: 2800, // Client-side timeout (slightly higher than server 2500ms)
    debug: false,
    userSync: {
      enabled: true,
      bidders: ['kargo', 'rubicon', 'pubmatic', 'sovrn', 'triplelift'],
      syncDelay: 1000,      // Delay before syncing (ms)
      iframeEnabled: true,  // Allow iframe syncs
      pixelEnabled: true,   // Allow pixel/redirect syncs
      maxSyncs: 5          // Max number of syncs per page
    }
  };

  // Active bid requests
  catalyst._bidRequests = {};
  catalyst._initialized = false;
  catalyst._userSyncComplete = false;
  catalyst._syncedBidders = [];

  // First-Party ID Manager
  catalyst._fpidManager = {
    cookieName: 'uids',  // Store in existing uids cookie
    fpidKey: 'fpid',
    expiryDays: 365,

    // Check if we have consent to generate/use FPID
    // GDPR compliance: Only allow FPID if consent is available or GDPR doesn't apply
    hasConsent: function() {
      // If no TCF API, assume GDPR doesn't apply (non-EU traffic)
      if (!window.__tcfapi) {
        return true;
      }

      // Check for cached consent data
      var hasValidConsent = false;
      var checkComplete = false;

      try {
        window.__tcfapi('getTCData', 2, function(tcData, success) {
          if (success && tcData) {
            // GDPR doesn't apply - allow FPID
            if (!tcData.gdprApplies) {
              hasValidConsent = true;
            }
            // GDPR applies - check for valid consent string
            else if (tcData.tcString && tcData.tcString.length >= 20) {
              hasValidConsent = true;
            }
          }
          checkComplete = true;
        });
      } catch (e) {
        catalyst.log('Error checking GDPR consent for FPID:', e);
        // On error, be conservative and deny FPID
        checkComplete = true;
      }

      // If TCF API is synchronous and completed, return result
      // Otherwise return false (no consent) to be safe
      return checkComplete ? hasValidConsent : false;
    },

    // Generate new FPID
    generate: function() {
      var timestamp = Date.now();
      var random = Math.random().toString(36).substr(2, 12);
      return 'fpi_' + timestamp + '_' + random;
    },

    // Get FPID from cookie (parse uids cookie JSON)
    get: function() {
      var uids = catalyst._getCookie(this.cookieName);
      if (uids) {
        try {
          var decoded = atob(uids);
          var data = JSON.parse(decoded);
          return data.fpid || null;
        } catch (e) {
          catalyst.log('Failed to parse fpid from cookie:', e);
          return null;
        }
      }
      return null;
    },

    // Generate or retrieve existing FPID (respects GDPR consent)
    getOrCreate: function() {
      // Always return existing FPID if available
      var fpid = this.get();
      if (fpid) {
        catalyst.log('Retrieved existing FPID:', fpid);
        return fpid;
      }

      // Only generate new FPID if we have consent
      if (!this.hasConsent()) {
        catalyst.log('FPID generation blocked - GDPR applies but no valid TCF consent');
        return null;
      }

      fpid = this.generate();
      catalyst.log('Generated new FPID:', fpid);
      return fpid;
    }
  };

  /**
   * Initialize Catalyst SDK
   * @param {Object} config - Configuration object
   * @param {string} config.accountId - MAI Publisher account ID
   * @param {string} [config.serverUrl] - Optional custom server URL
   * @param {number} [config.timeout] - Optional timeout in ms (default: 2800)
   * @param {boolean} [config.debug] - Enable debug logging
   * @param {Function} [callback] - Optional callback function called when SDK is ready (after user sync completes)
   */
  catalyst.init = function(config, callback) {
    if (!config || !config.accountId) {
      catalyst.log('Error: accountId is required');
      if (typeof callback === 'function') {
        callback();
      }
      return;
    }

    catalyst._config.accountId = config.accountId;

    if (config.serverUrl) {
      catalyst._config.serverUrl = config.serverUrl;
    }

    if (config.timeout) {
      catalyst._config.timeout = config.timeout;
    }

    if (config.debug !== undefined) {
      catalyst._config.debug = config.debug;
    }

    // Allow override of user sync settings
    if (config.userSync !== undefined) {
      if (typeof config.userSync === 'boolean') {
        catalyst._config.userSync.enabled = config.userSync;
      } else if (typeof config.userSync === 'object') {
        for (var key in config.userSync) {
          if (config.userSync.hasOwnProperty(key)) {
            catalyst._config.userSync[key] = config.userSync[key];
          }
        }
      }
    }

    catalyst._initialized = true;
    catalyst.log('Catalyst SDK initialized with accountId:', config.accountId);

    // Trigger user sync IMMEDIATELY (no delay) and wait for completion before callback
    if (catalyst._config.userSync.enabled) {
      catalyst._performUserSync(function() {
        catalyst.log('User sync complete - SDK ready for bid requests');
        if (typeof callback === 'function') {
          callback();
        }
      });
    } else {
      // User sync disabled, ready immediately
      if (typeof callback === 'function') {
        callback();
      }
    }
  };

  /**
   * Request bids for ad slots
   * @param {Object} config - Bid request configuration
   * @param {Array} config.slots - Array of ad slot configurations
   * @param {Function} callback - Callback function(bids) called when bids are ready or timeout
   */
  catalyst.requestBids = function(config, callback) {
    if (!catalyst._initialized) {
      catalyst.log('Error: Catalyst SDK not initialized. Call catalyst.init() first.');
      if (typeof callback === 'function') {
        callback([]);
      }
      return;
    }

    // Warn if user sync hasn't completed yet (may result in lower bid responses)
    if (catalyst._config.userSync.enabled && !catalyst._userSyncComplete) {
      catalyst.log('Warning: requestBids called before user sync completed - bidders may not have synced IDs yet');
    }

    if (!config || !config.slots || config.slots.length === 0) {
      catalyst.log('Error: slots array is required');
      if (typeof callback === 'function') {
        callback([]);
      }
      return;
    }

    var requestId = 'catalyst-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    var startTime = Date.now();
    var timeoutMs = catalyst._config.timeout;

    // Build MAI bid request
    var bidRequest = {
      accountId: catalyst._config.accountId,
      timeout: timeoutMs,
      slots: [],
      page: {
        url: window.location.href,
        domain: window.location.hostname,
        keywords: [],
        categories: []
      },
      device: {
        width: window.screen.width,
        height: window.screen.height,
        deviceType: catalyst._detectDeviceType(),
        userAgent: navigator.userAgent
      }
    };

    // Process slots
    for (var i = 0; i < config.slots.length; i++) {
      var slot = config.slots[i];

      if (!slot.divId || !slot.sizes || slot.sizes.length === 0) {
        catalyst.log('Warning: Invalid slot configuration, skipping:', slot);
        continue;
      }

      // Skip out-of-page slots (special GPT formats like interstitials, overlays)
      if (Array.isArray(slot.sizes) && slot.sizes.length === 1 &&
          typeof slot.sizes[0] === 'string' && slot.sizes[0].toLowerCase() === 'out-of-page') {
        catalyst.log('Skipping out-of-page slot (not eligible for bidding):', slot.divId);
        continue;
      }

      // Normalize sizes to [[w, h], ...] format
      var normalizedSizes = catalyst._normalizeSizes(slot.sizes);
      if (!normalizedSizes || normalizedSizes.length === 0) {
        catalyst.log('Warning: Could not normalize sizes for slot:', slot.divId,
                     '| Raw sizes:', JSON.stringify(slot.sizes),
                     '| Type:', typeof slot.sizes,
                     '| Full slot:', JSON.stringify(slot));
        continue;
      }

      bidRequest.slots.push({
        divId: slot.divId,
        sizes: normalizedSizes,
        adUnitPath: slot.adUnitPath || '',
        position: slot.position || '',
        enabled_bidders: slot.enabled_bidders || ['catalyst']
      });
    }

    // Early exit if no eligible slots - prevents empty bid requests
    if (bidRequest.slots.length === 0) {
      catalyst.log('No eligible slots found - skipping bid request (all slots filtered out or not ready)');
      if (typeof callback === 'function') {
        callback([]);
      }
      return;
    }

    catalyst.log('Requesting bids for', bidRequest.slots.length, 'slots with timeout', timeoutMs + 'ms');

    // Add page context if provided
    if (config.page) {
      if (config.page.keywords) {
        bidRequest.page.keywords = config.page.keywords;
      }
      if (config.page.categories) {
        bidRequest.page.categories = config.page.categories;
      }
    }

    // Include user IDs from cookie
    var userIds = catalyst._getUserIds();
    if (Object.keys(userIds).length > 0) {
      bidRequest.user = bidRequest.user || {};
      bidRequest.user.userIds = userIds;
      catalyst.log('Including user IDs for bidders:', Object.keys(userIds).join(', '));
    }

    // Include FPID (only if consent is available)
    var fpid = catalyst._fpidManager.getOrCreate();
    if (fpid) {
      bidRequest.user = bidRequest.user || {};
      bidRequest.user.fpid = fpid;
      catalyst.log('Including FPID:', fpid);
    } else {
      catalyst.log('FPID not included - consent not available');
    }

    // Collect ORTB2 data from Prebid.js
    var ortb2Data = catalyst._getORTB2Data();

    // Merge ORTB2 site data (content categories, keywords, etc.)
    if (ortb2Data.site) {
      if (ortb2Data.site.cat && ortb2Data.site.cat.length > 0) {
        bidRequest.page.categories = ortb2Data.site.cat;
      }
      if (ortb2Data.site.keywords) {
        bidRequest.page.keywords = ortb2Data.site.keywords;
      }
      if (ortb2Data.site.content && Object.keys(ortb2Data.site.content).length > 0) {
        bidRequest.page.content = ortb2Data.site.content;
      }
      if (ortb2Data.site.ext && Object.keys(ortb2Data.site.ext).length > 0) {
        bidRequest.page.ext = ortb2Data.site.ext;
      }
    }

    // Merge ORTB2 device data (geo targeting, connection info)
    if (ortb2Data.device) {
      if (ortb2Data.device.geo && Object.keys(ortb2Data.device.geo).length > 0) {
        bidRequest.device.geo = ortb2Data.device.geo;
      }
      if (ortb2Data.device.connectiontype) {
        bidRequest.device.connectiontype = ortb2Data.device.connectiontype;
      }
      if (ortb2Data.device.ext && Object.keys(ortb2Data.device.ext).length > 0) {
        bidRequest.device.ext = ortb2Data.device.ext;
      }
    }

    // Merge ORTB2 user data (segments, interests, demographics)
    if (ortb2Data.user) {
      bidRequest.user = bidRequest.user || {};
      if (ortb2Data.user.data && ortb2Data.user.data.length > 0) {
        bidRequest.user.data = ortb2Data.user.data;
      }
      if (ortb2Data.user.ext && Object.keys(ortb2Data.user.ext).length > 0) {
        bidRequest.user.ext = bidRequest.user.ext || {};
        for (var key in ortb2Data.user.ext) {
          if (ortb2Data.user.ext.hasOwnProperty(key)) {
            bidRequest.user.ext[key] = ortb2Data.user.ext[key];
          }
        }
      }
    }

    // Setup timeout
    var timedOut = false;
    var timeoutId = setTimeout(function() {
      timedOut = true;
      catalyst.log('Bid request timed out after', timeoutMs + 'ms');

      if (typeof callback === 'function') {
        callback([]);
      }

      delete catalyst._bidRequests[requestId];
    }, timeoutMs);

    // Store bid request
    catalyst._bidRequests[requestId] = {
      config: config,
      callback: callback,
      startTime: startTime,
      timeoutId: timeoutId
    };

    // Get privacy consent data, then send bid request
    var sendBidRequest = function() {
      // Make POST request to /v1/bid endpoint
      catalyst._makeBidRequest(bidRequest, function(error, response) {
      if (timedOut) {
        catalyst.log('Ignoring response after timeout');
        return;
      }

      clearTimeout(timeoutId);
      delete catalyst._bidRequests[requestId];

      var elapsedMs = Date.now() - startTime;

      if (error) {
        catalyst.log('Bid request failed:', error, 'in', elapsedMs + 'ms');
        if (typeof callback === 'function') {
          callback([]);
        }
        return;
      }

      var bids = response.bids || [];
      catalyst.log('Received', bids.length, 'bids in', elapsedMs + 'ms');

      if (typeof callback === 'function') {
        callback(bids);
      }
      });
    };

    // Add privacy/consent info if available, then send request
    if (window.__tcfapi || window.__uspapi || window.__cmp) {
      catalyst._addPrivacyConsent(bidRequest, sendBidRequest);
    } else {
      // No privacy APIs available, send immediately
      sendBidRequest();
    }
  };

  /**
   * Make POST request to bid endpoint
   * @param {Object} bidRequest - Bid request payload
   * @param {Function} callback - Callback function(error, response)
   * @private
   */
  catalyst._makeBidRequest = function(bidRequest, callback) {
    var url = catalyst._config.serverUrl + '/v1/bid';

    // Log complete bid request if debug enabled
    if (catalyst._config.debug) {
      catalyst.log('=== FULL BID REQUEST ===');
      catalyst.log(JSON.stringify(bidRequest, null, 2));
      catalyst.log('========================');
    }

    var xhr = new XMLHttpRequest();
    xhr.open('POST', url, true);
    xhr.setRequestHeader('Content-Type', 'application/json');
    xhr.timeout = catalyst._config.timeout;

    xhr.onload = function() {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          var response = JSON.parse(xhr.responseText);

          // Log complete bid response if debug enabled
          if (catalyst._config.debug) {
            catalyst.log('=== FULL BID RESPONSE ===');
            catalyst.log(JSON.stringify(response, null, 2));
            catalyst.log('=========================');
          }

          callback(null, response);
        } catch (e) {
          catalyst.log('Error parsing response:', e);
          catalyst.log('Response text:', xhr.responseText);
          callback(e, null);
        }
      } else {
        catalyst.log('Bid request failed with status:', xhr.status);
        catalyst.log('Response:', xhr.responseText);
        callback(new Error('HTTP ' + xhr.status), null);
      }
    };

    xhr.onerror = function() {
      catalyst.log('Network error making bid request');
      callback(new Error('Network error'), null);
    };

    xhr.ontimeout = function() {
      catalyst.log('Bid request timeout after', catalyst._config.timeout, 'ms');
      callback(new Error('Request timeout'), null);
    };

    try {
      xhr.send(JSON.stringify(bidRequest));
    } catch (e) {
      catalyst.log('Exception sending bid request:', e);
      callback(e, null);
    }
  };

  /**
   * Notify MAI Publisher that bids are ready
   * @private
   * @deprecated This function is no longer used. Catalyst now relies on callbacks only.
   */
  catalyst._notifyReady = function() {
    // No-op: Callback pattern is now used instead of global function calls
    // This function is kept for backwards compatibility but does nothing
  };

  /**
   * Read and parse the uids cookie AND get user IDs from Prebid.js
   * @returns {Object} Map of bidder -> user ID
   * @private
   */
  catalyst._getUserIds = function() {
    var userIds = {};

    // 1. Read from Catalyst's own uids cookie
    var uids = catalyst._getCookie('uids');
    if (uids) {
      try {
        // Cookie is base64-encoded JSON
        var decoded = atob(uids);
        var data = JSON.parse(decoded);

        // Extract just the UID values (strip expires timestamps)
        for (var bidder in data.uids || {}) {
          if (data.uids[bidder].uid && !catalyst._isExpired(data.uids[bidder].expires)) {
            userIds[bidder] = data.uids[bidder].uid;
          }
        }
      } catch (e) {
        catalyst.log('Failed to parse uids cookie:', e);
      }
    }

    // 2. Get user IDs from Prebid.js if available
    if (window.pbjs && typeof window.pbjs.getUserIds === 'function') {
      try {
        var prebidUserIds = window.pbjs.getUserIds();
        if (prebidUserIds && typeof prebidUserIds === 'object') {
          // Prebid returns IDs like: { id5id: {...}, pubcid: "...", ... }
          // Map known ID sources to bidder codes
          var idSourceToBidder = {
            'id5id': 'id5',
            'pubcid': 'pubcommon',
            'pubProvidedId': 'pubprovided',
            'uid2': 'uid2',
            'parrableId': 'parrable',
            'identityLink': 'liveramp',
            'criteoId': 'criteo',
            'netId': 'netid'
          };

          for (var source in prebidUserIds) {
            var bidderCode = idSourceToBidder[source] || source;
            var idValue = prebidUserIds[source];

            // Handle different ID formats
            if (idValue && typeof idValue === 'object') {
              if (idValue.uid) {
                userIds[bidderCode] = idValue.uid;
              } else if (idValue.id) {
                userIds[bidderCode] = idValue.id;
              }
            } else if (idValue && typeof idValue === 'string') {
              userIds[bidderCode] = idValue;
            }
          }

          if (Object.keys(prebidUserIds).length > 0) {
            catalyst.log('Added', Object.keys(prebidUserIds).length, 'user IDs from Prebid.js');
          }
        }
      } catch (e) {
        catalyst.log('Failed to get Prebid user IDs:', e);
      }
    }

    return userIds;
  };

  /**
   * Get ORTB2 data from Prebid.js configuration
   * This collects site, device, and user data for enhanced targeting
   * @returns {Object} ORTB2 data with site, device, and user properties
   * @private
   */
  catalyst._getORTB2Data = function() {
    var ortb2 = {
      site: null,
      device: null,
      user: null
    };

    if (window.pbjs && typeof window.pbjs.getConfig === 'function') {
      try {
        // Get ORTB2 site data (content categories, keywords, publisher FPD)
        var siteData = window.pbjs.getConfig('ortb2.site');
        if (siteData) {
          ortb2.site = {
            cat: siteData.cat || [],           // IAB content categories
            keywords: siteData.keywords || '', // Page keywords
            content: siteData.content || {},   // Content metadata
            ext: siteData.ext || {}            // Publisher first-party data
          };
          catalyst.log('Added ORTB2 site data with', ortb2.site.cat.length, 'categories');
        }

        // Get ORTB2 device data (geo targeting, connection info)
        var deviceData = window.pbjs.getConfig('ortb2.device');
        if (deviceData) {
          ortb2.device = {
            geo: deviceData.geo || {},         // Geographic targeting data
            connectiontype: deviceData.connectiontype,
            ext: deviceData.ext || {}
          };
          if (ortb2.device.geo.country) {
            catalyst.log('Added ORTB2 device data with geo:', ortb2.device.geo.country);
          }
        }

        // Get ORTB2 user data (segments, interests, demographics)
        var userData = window.pbjs.getConfig('ortb2.user');
        if (userData) {
          ortb2.user = {
            data: userData.data || [],         // User segments and interests
            ext: userData.ext || {}            // User first-party data
          };
          if (ortb2.user.data.length > 0) {
            catalyst.log('Added ORTB2 user data with', ortb2.user.data.length, 'segments');
          }
        }
      } catch (e) {
        catalyst.log('Failed to get ORTB2 data from Prebid.js:', e);
      }
    }

    return ortb2;
  };

  /**
   * Get a cookie by name
   * @param {string} name - Cookie name
   * @returns {string|null} Cookie value or null
   * @private
   */
  catalyst._getCookie = function(name) {
    var cookies = document.cookie.split(';');
    for (var i = 0; i < cookies.length; i++) {
      var cookie = cookies[i].trim();
      if (cookie.indexOf(name + '=') === 0) {
        return cookie.substring(name.length + 1);
      }
    }
    return null;
  };

  /**
   * Check if a timestamp has expired
   * @param {string} expiresStr - ISO date string
   * @returns {boolean} True if expired
   * @private
   */
  catalyst._isExpired = function(expiresStr) {
    if (!expiresStr) return true;
    try {
      return new Date(expiresStr) < new Date();
    } catch (e) {
      return true;
    }
  };

  /**
   * Add privacy consent to bid request
   * @param {Object} bidRequest - Bid request object
   * @param {Function} callback - Called when consent data is ready
   * @private
   */
  catalyst._addPrivacyConsent = function(bidRequest, callback) {
    bidRequest.user = bidRequest.user || {};

    var tcfDone = false;
    var uspDone = false;
    var timeout = 200; // Max 200ms wait for consent data
    var timeoutId = null;

    var checkComplete = function() {
      if ((tcfDone || !window.__tcfapi) && (uspDone || !window.__uspapi)) {
        if (timeoutId) clearTimeout(timeoutId);
        if (callback) callback();
      }
    };

    // Timeout fallback - don't wait forever for consent data
    timeoutId = setTimeout(function() {
      catalyst.log('Privacy consent timeout, proceeding without full data');
      tcfDone = true;
      uspDone = true;
      if (callback) callback();
    }, timeout);

    // Try to get GDPR consent via TCF API
    if (window.__tcfapi) {
      try {
        window.__tcfapi('getTCData', 2, function(tcData, success) {
          tcfDone = true;
          if (success && tcData) {
            bidRequest.user.gdprApplies = tcData.gdprApplies || false;
            bidRequest.user.consentGiven = tcData.eventStatus === 'tcloaded' || tcData.eventStatus === 'useractioncomplete';

            // CRITICAL: Pass the actual TCF consent string for bidders (only if GDPR applies)
            if (tcData.gdprApplies && tcData.tcString) {
              bidRequest.user.consentString = tcData.tcString;
              catalyst.log('Added TCF consent string for GDPR region');
            }
          }
          checkComplete();
        });
      } catch (e) {
        catalyst.log('Error getting GDPR consent:', e);
        tcfDone = true;
        checkComplete();
      }
    } else {
      tcfDone = true;
    }

    // Try to get US Privacy string
    if (window.__uspapi) {
      try {
        window.__uspapi('getUSPData', 1, function(uspData, success) {
          uspDone = true;
          if (success && uspData && uspData.uspString) {
            bidRequest.user.uspConsent = uspData.uspString;
          } else {
            // CMP present but no USP string - use permissive default for testing
            bidRequest.user.uspConsent = '1YNN';
          }
          checkComplete();
        });
      } catch (e) {
        catalyst.log('Error getting USP consent:', e);
        // Use permissive default on error
        bidRequest.user.uspConsent = '1YNN';
        uspDone = true;
        checkComplete();
      }
    } else {
      // No CMP - use permissive default for testing
      // TODO PRODUCTION: Remove this default and require proper CMP
      bidRequest.user.uspConsent = '1YNN';
      uspDone = true;
    }

    // If neither API is available, call back immediately
    if (!window.__tcfapi && !window.__uspapi) {
      checkComplete();
    }
  };

  /**
   * Perform user sync with configured bidders
   * @private
   */
  catalyst._performUserSync = function(onComplete) {
    if (!catalyst._config.userSync.enabled) {
      catalyst.log('User sync disabled');
      if (typeof onComplete === 'function') {
        onComplete();
      }
      return;
    }

    if (catalyst._userSyncComplete) {
      catalyst.log('User sync already performed');
      if (typeof onComplete === 'function') {
        onComplete();
      }
      return;
    }

    // Privacy consent will be checked when we gather privacy data for the sync request
    // Server also validates consent as a backup (defense in depth)
    catalyst.log('Starting user sync for bidders:', catalyst._config.userSync.bidders);

    // Build cookie sync request
    var syncRequest = {
      bidders: catalyst._config.userSync.bidders,
      gdpr: 0,
      gdpr_consent: '',
      us_privacy: '',
      limit: catalyst._config.userSync.maxSyncs
    };

    // Include FPID only if consent is available
    var fpid = catalyst._fpidManager.getOrCreate();
    if (fpid) {
      syncRequest.fpid = fpid;
      catalyst.log('Including FPID in cookie sync:', fpid);
    } else {
      catalyst.log('FPID not included in cookie sync - consent not available');
    }

    // Function to send cookie sync request
    var sendSyncRequest = function() {
      var url = catalyst._config.serverUrl + '/cookie_sync';
      var xhr = new XMLHttpRequest();
      xhr.open('POST', url, true);
      xhr.setRequestHeader('Content-Type', 'application/json');
      xhr.timeout = 5000;

      xhr.onload = function() {
        if (xhr.status >= 200 && xhr.status < 300) {
          try {
            var response = JSON.parse(xhr.responseText);
            catalyst._fireSyncPixels(response);
          } catch (e) {
            catalyst.log('Error parsing sync response:', e);
          }
        } else {
          catalyst.log('User sync request failed:', xhr.status);
        }
        // Call completion callback even on failure
        if (typeof onComplete === 'function') {
          onComplete();
        }
      };

      xhr.onerror = function() {
        catalyst.log('User sync network error');
        // Call completion callback even on error
        if (typeof onComplete === 'function') {
          onComplete();
        }
      };

      xhr.ontimeout = function() {
        catalyst.log('User sync timeout');
        // Call completion callback even on timeout
        if (typeof onComplete === 'function') {
          onComplete();
        }
      };

      try {
        catalyst.log('Sending cookie sync with privacy params:', {
          gdpr: syncRequest.gdpr,
          gdpr_consent: syncRequest.gdpr_consent ? syncRequest.gdpr_consent.substring(0, 20) + '...' : '',
          us_privacy: syncRequest.us_privacy
        });
        xhr.send(JSON.stringify(syncRequest));
      } catch (e) {
        catalyst.log('Error sending sync request:', e);
        // Call completion callback even on exception
        if (typeof onComplete === 'function') {
          onComplete();
        }
      }

      catalyst._userSyncComplete = true;
    };

    // Add privacy parameters if available, then send request
    if (window.__tcfapi || window.__uspapi) {
      catalyst._addPrivacyToSyncRequest(syncRequest, sendSyncRequest);
    } else {
      // No privacy APIs available, send immediately
      sendSyncRequest();
    }
  };

  /**
   * Fire sync pixels/iframes for bidders
   * @param {Object} response - Cookie sync response
   * @private
   */
  catalyst._fireSyncPixels = function(response) {
    if (!response.bidder_status || response.bidder_status.length === 0) {
      catalyst.log('No sync URLs to fire');
      return;
    }

    var config = catalyst._config.userSync;
    var syncsFired = 0;

    for (var i = 0; i < response.bidder_status.length; i++) {
      var bidderSync = response.bidder_status[i];

      if (!bidderSync.usersync || !bidderSync.usersync.url) {
        continue;
      }

      var syncInfo = bidderSync.usersync;
      var syncType = syncInfo.type;

      // Check if sync type is enabled
      if (syncType === 'iframe' && !config.iframeEnabled) {
        catalyst.log('Skipping iframe sync for', syncInfo.bidder);
        continue;
      }

      if (syncType === 'redirect' && !config.pixelEnabled) {
        catalyst.log('Skipping pixel sync for', syncInfo.bidder);
        continue;
      }

      // Fire the sync
      if (syncType === 'iframe') {
        catalyst._fireIframeSync(syncInfo.url, syncInfo.bidder);
      } else {
        catalyst._firePixelSync(syncInfo.url, syncInfo.bidder);
      }

      catalyst._syncedBidders.push(syncInfo.bidder);
      syncsFired++;

      if (syncsFired >= config.maxSyncs) {
        catalyst.log('Max syncs reached:', config.maxSyncs);
        break;
      }
    }

    catalyst.log('Fired', syncsFired, 'user syncs');
  };

  /**
   * Fire iframe sync
   * @param {string} url - Sync URL
   * @param {string} bidder - Bidder code
   * @private
   */
  catalyst._fireIframeSync = function(url, bidder) {
    try {
      var iframe = document.createElement('iframe');
      iframe.src = url;
      iframe.style.display = 'none';
      iframe.width = 0;
      iframe.height = 0;
      iframe.setAttribute('data-bidder', bidder);
      document.body.appendChild(iframe);
      catalyst.log('Fired iframe sync for', bidder);

      // Remove iframe after 10 seconds to prevent memory leak
      setTimeout(function() {
        try {
          if (iframe.parentNode) {
            iframe.parentNode.removeChild(iframe);
            catalyst.log('Cleaned up iframe sync for', bidder);
          }
        } catch (e) {
          catalyst.log('Error cleaning up iframe sync for', bidder, e);
        }
      }, 10000);
    } catch (e) {
      catalyst.log('Error firing iframe sync for', bidder, e);
    }
  };

  /**
   * Fire pixel/redirect sync
   * @param {string} url - Sync URL
   * @param {string} bidder - Bidder code
   * @private
   */
  catalyst._firePixelSync = function(url, bidder) {
    try {
      var img = new Image();
      img.src = url;
      img.style.display = 'none';
      img.width = 1;
      img.height = 1;
      img.setAttribute('data-bidder', bidder);
      catalyst.log('Fired pixel sync for', bidder);
    } catch (e) {
      catalyst.log('Error firing pixel sync for', bidder, e);
    }
  };

  /**
   * Check if user has given privacy consent for syncing
   * @returns {boolean} True if consent given or not required
   * @private
   * @deprecated This function has async race conditions and is no longer used
   *   Server validates consent properly with actual privacy data
   */
  catalyst._hasPrivacyConsent = function() {
    // Kept for backwards compatibility only
    // Actual consent validation happens server-side with proper privacy strings
    // collected asynchronously via _addPrivacyConsent() and _addPrivacyToSyncRequest()
    return true;
  };

  /**
   * Add privacy parameters to sync request
   * @param {Object} syncRequest - Sync request object to modify
   * @param {Function} callback - Called when privacy data is ready
   * @private
   */
  catalyst._addPrivacyToSyncRequest = function(syncRequest, callback) {
    var tcfDone = false;
    var uspDone = false;
    var timeout = 200; // Max 200ms wait for consent data
    var timeoutId = null;

    var checkComplete = function() {
      if ((tcfDone || !window.__tcfapi) && (uspDone || !window.__uspapi)) {
        if (timeoutId) clearTimeout(timeoutId);
        if (callback) callback();
      }
    };

    // Timeout fallback - don't wait forever for consent data
    timeoutId = setTimeout(function() {
      catalyst.log('Cookie sync privacy timeout, proceeding');
      tcfDone = true;
      uspDone = true;
      if (callback) callback();
    }, timeout);

    // Try to get GDPR consent
    if (window.__tcfapi) {
      try {
        window.__tcfapi('getTCData', 2, function(tcData, success) {
          tcfDone = true;
          if (success && tcData) {
            syncRequest.gdpr = tcData.gdprApplies ? 1 : 0;
            syncRequest.gdpr_consent = tcData.tcString || '';
            if (tcData.gdprApplies && tcData.tcString) {
              catalyst.log('Added TCF consent string for cookie sync');
            }
          }
          checkComplete();
        });
      } catch (e) {
        catalyst.log('Error getting GDPR consent for sync:', e);
        tcfDone = true;
        checkComplete();
      }
    } else {
      tcfDone = true;
    }

    // Try to get US Privacy string
    if (window.__uspapi) {
      try {
        window.__uspapi('getUSPData', 1, function(uspData, success) {
          uspDone = true;
          if (success && uspData && uspData.uspString) {
            syncRequest.us_privacy = uspData.uspString;
            catalyst.log('Added USP consent string for cookie sync');
          } else {
            // CMP present but no USP string - use permissive default for testing
            syncRequest.us_privacy = '1YNN';
            catalyst.log('Using default USP string (CMP present but no data)');
          }
          checkComplete();
        });
      } catch (e) {
        catalyst.log('Error getting USP consent for sync:', e);
        // Use permissive default on error
        syncRequest.us_privacy = '1YNN';
        uspDone = true;
        checkComplete();
      }
    } else {
      // No CMP - use permissive default for testing
      // TODO PRODUCTION: Remove this default and require proper CMP
      syncRequest.us_privacy = '1YNN';
      catalyst.log('Using default USP string (no CMP found)');
      uspDone = true;
    }

    // If neither API is available, call back immediately
    if (!window.__tcfapi && !window.__uspapi) {
      checkComplete();
    }
  };

  /**
   * Normalize sizes to [[width, height], ...] format
   * Handles multiple input formats:
   * - [[300, 250], [728, 90]] -> [[300, 250], [728, 90]] (passthrough)
   * - [300, 250] -> [[300, 250]]
   * - "300x250" -> [[300, 250]]
   * - ["300x250", "728x90"] -> [[300, 250], [728, 90]]
   * @param {*} sizes - Sizes in any format
   * @returns {Array} Normalized sizes as [[w, h], ...]
   * @private
   */
  catalyst._normalizeSizes = function(sizes) {
    if (!sizes) {
      catalyst.log('DEBUG: _normalizeSizes received null/undefined');
      return [];
    }

    var normalized = [];

    // Handle array input
    if (Array.isArray(sizes)) {
      for (var i = 0; i < sizes.length; i++) {
        var size = sizes[i];

        // Skip null/undefined entries
        if (size == null) {
          catalyst.log('DEBUG: Skipping null/undefined size at index', i);
          continue;
        }

        // Already in [width, height] format
        if (Array.isArray(size) && size.length === 2 &&
            typeof size[0] === 'number' && typeof size[1] === 'number' &&
            size[0] > 0 && size[1] > 0) {
          normalized.push(size);
        }
        // Try to coerce array with string numbers: ["300", "250"]
        else if (Array.isArray(size) && size.length === 2) {
          var w = parseInt(size[0], 10);
          var h = parseInt(size[1], 10);
          if (!isNaN(w) && !isNaN(h) && w > 0 && h > 0) {
            normalized.push([w, h]);
            catalyst.log('DEBUG: Coerced string sizes to numbers:', [w, h]);
          } else {
            catalyst.log('DEBUG: Could not coerce array size:', JSON.stringify(size));
          }
        }
        // String format: "300x250"
        else if (typeof size === 'string') {
          var parsed = catalyst._parseSizeString(size);
          if (parsed) {
            normalized.push(parsed);
          } else {
            catalyst.log('DEBUG: Could not parse size string:', size);
          }
        }
        // Single [width, height] - not nested
        else if (i === 0 && typeof size === 'number' && typeof sizes[1] === 'number' &&
                 size > 0 && sizes[1] > 0) {
          // Input is [300, 250] not [[300, 250]]
          normalized.push([sizes[0], sizes[1]]);
          break; // Done processing
        }
        // Object format: {width: 300, height: 250} or {w: 300, h: 250}
        else if (typeof size === 'object') {
          var w = size.width || size.w;
          var h = size.height || size.h;
          if (typeof w === 'number' && typeof h === 'number' && w > 0 && h > 0) {
            normalized.push([w, h]);
            catalyst.log('DEBUG: Converted object size to array:', [w, h]);
          } else {
            catalyst.log('DEBUG: Invalid object size format:', JSON.stringify(size));
          }
        }
        else {
          catalyst.log('DEBUG: Unrecognized size format at index', i, ':', JSON.stringify(size), 'type:', typeof size);
        }
      }
    }
    // Handle string input: "300x250"
    else if (typeof sizes === 'string') {
      var parsed = catalyst._parseSizeString(sizes);
      if (parsed) {
        normalized.push(parsed);
      } else {
        catalyst.log('DEBUG: Could not parse sizes string:', sizes);
      }
    }
    // Handle object input: {width: 300, height: 250}
    else if (typeof sizes === 'object') {
      var w = sizes.width || sizes.w;
      var h = sizes.height || sizes.h;
      if (typeof w === 'number' && typeof h === 'number' && w > 0 && h > 0) {
        normalized.push([w, h]);
        catalyst.log('DEBUG: Converted object sizes to array:', [w, h]);
      } else {
        catalyst.log('DEBUG: Invalid object sizes format:', JSON.stringify(sizes));
      }
    }
    else {
      catalyst.log('DEBUG: Unrecognized sizes type:', typeof sizes, 'value:', sizes);
    }

    if (normalized.length === 0) {
      catalyst.log('DEBUG: _normalizeSizes returning empty array. Input was:', JSON.stringify(sizes));
    }

    return normalized;
  };

  /**
   * Parse size string like "300x250" to [300, 250]
   * @param {string} sizeStr - Size string
   * @returns {Array|null} [width, height] or null if invalid
   * @private
   */
  catalyst._parseSizeString = function(sizeStr) {
    if (typeof sizeStr !== 'string') {
      return null;
    }

    var parts = sizeStr.split('x');
    if (parts.length !== 2) {
      return null;
    }

    var width = parseInt(parts[0], 10);
    var height = parseInt(parts[1], 10);

    if (isNaN(width) || isNaN(height) || width <= 0 || height <= 0) {
      return null;
    }

    return [width, height];
  };

  /**
   * Detect device type
   * @returns {string} Device type: 'mobile', 'tablet', or 'desktop'
   * @private
   */
  catalyst._detectDeviceType = function() {
    var ua = navigator.userAgent;

    if (/(tablet|ipad|playbook|silk)|(android(?!.*mobi))/i.test(ua)) {
      return 'tablet';
    }

    if (/Mobile|Android|iP(hone|od)|IEMobile|BlackBerry|Kindle|Silk-Accelerated|(hpw|web)OS|Opera M(obi|ini)/.test(ua)) {
      return 'mobile';
    }

    return 'desktop';
  };

  /**
   * Set targeting for Google Publisher Tag (GPT)
   * Sets Catalyst bid data as targeting key-value pairs for GAM
   * @param {Object|Array} targetingData - Targeting data or array of bids
   */
  catalyst.setTargeting = function(targetingData) {
    if (!window.googletag || !window.googletag.pubads) {
      catalyst.log('Warning: googletag not available for setTargeting');
      return;
    }

    try {
      var pubads = window.googletag.pubads();

      // Handle different input formats
      if (Array.isArray(targetingData)) {
        // Array of bids - set targeting for each
        for (var i = 0; i < targetingData.length; i++) {
          var bid = targetingData[i];
          if (bid && bid.divId) {
            catalyst._setSlotTargeting(bid);
          }
        }
      } else if (targetingData && typeof targetingData === 'object') {
        // Single bid object or targeting object
        if (targetingData.divId) {
          catalyst._setSlotTargeting(targetingData);
        } else {
          // Key-value pairs object
          for (var key in targetingData) {
            if (targetingData.hasOwnProperty(key)) {
              pubads.setTargeting(key, targetingData[key]);
              catalyst.log('Set GPT targeting:', key, '=', targetingData[key]);
            }
          }
        }
      }
    } catch (e) {
      catalyst.log('Error setting GPT targeting:', e);
    }
  };

  /**
   * Set targeting for a specific slot
   * @param {Object} bid - Bid object with targeting data
   * @private
   */
  catalyst._setSlotTargeting = function(bid) {
    if (!bid || !bid.divId) {
      return;
    }

    try {
      var pubads = window.googletag.pubads();

      // Find the GPT slot for this divId
      var slots = window.googletag.pubads().getSlots();
      var targetSlot = null;

      for (var i = 0; i < slots.length; i++) {
        var slotElementId = slots[i].getSlotElementId();
        if (slotElementId === bid.divId || slotElementId === 'mai-ad-' + bid.divId) {
          targetSlot = slots[i];
          break;
        }
      }

      if (targetSlot) {
        // Set Catalyst-specific header bidding keys (no overlap with Prebid)
        if (bid.cpm) {
          targetSlot.setTargeting('hb_pb_catalyst', bid.cpm.toFixed(2));
        }

        if (bid.creativeId) {
          targetSlot.setTargeting('hb_adid_catalyst', bid.creativeId);
          targetSlot.setTargeting('hb_creative_catalyst', bid.creativeId);
        }

        if (bid.width && bid.height) {
          targetSlot.setTargeting('hb_size_catalyst', bid.width + 'x' + bid.height);
        }

        // Set bid source (server-to-server)
        targetSlot.setTargeting('hb_source_catalyst', 's2s');

        // Set format (banner ads)
        targetSlot.setTargeting('hb_format_catalyst', 'banner');

        // Set deal ID if present (for PMP deals)
        if (bid.dealId) {
          targetSlot.setTargeting('hb_deal_catalyst', bid.dealId);
        }

        // Set advertiser domain if available
        if (bid.meta && bid.meta.advertiserDomains && bid.meta.advertiserDomains.length > 0) {
          targetSlot.setTargeting('hb_adomain_catalyst', bid.meta.advertiserDomains[0]);
        }

        // Set actual demand partner that won (if available in meta)
        if (bid.meta && bid.meta.networkName) {
          targetSlot.setTargeting('hb_partner', bid.meta.networkName);
          targetSlot.setTargeting('hb_bidder_catalyst', bid.meta.networkName);
          catalyst.log('Set slot targeting for', bid.divId, 'CPM:', bid.cpm, 'Partner:', bid.meta.networkName);
        } else {
          catalyst.log('Set slot targeting for', bid.divId, 'CPM:', bid.cpm);
        }
      } else {
        // Set page-level targeting if no slot found (Catalyst-specific keys only)
        if (bid.cpm) {
          pubads.setTargeting('hb_pb_catalyst', bid.cpm.toFixed(2));
        }

        // Set bid source and format
        pubads.setTargeting('hb_source_catalyst', 's2s');
        pubads.setTargeting('hb_format_catalyst', 'banner');

        // Set deal ID if present
        if (bid.dealId) {
          pubads.setTargeting('hb_deal_catalyst', bid.dealId);
        }

        // Set advertiser domain if available
        if (bid.meta && bid.meta.advertiserDomains && bid.meta.advertiserDomains.length > 0) {
          pubads.setTargeting('hb_adomain_catalyst', bid.meta.advertiserDomains[0]);
        }

        // Set actual demand partner
        if (bid.meta && bid.meta.networkName) {
          pubads.setTargeting('hb_partner', bid.meta.networkName);
          pubads.setTargeting('hb_bidder_catalyst', bid.meta.networkName);
        }

        catalyst.log('Set page-level targeting (slot not found):', bid.divId);
      }
    } catch (e) {
      catalyst.log('Error setting slot targeting:', e);
    }
  };

  /**
   * Log message (if debug enabled)
   * @param {...*} args - Arguments to log
   */
  catalyst.log = function() {
    if (catalyst._config.debug && console && console.log) {
      console.log.apply(console, ['[Catalyst]'].concat(Array.prototype.slice.call(arguments)));
    }
  };

  /**
   * Get SDK version
   * @returns {string} Version string
   */
  catalyst.version = function() {
    return '1.0.0';
  };

  // Process command queue
  catalyst._processCommandQueue = function() {
    while (catalyst.cmd.length > 0) {
      var cmd = catalyst.cmd.shift();
      if (typeof cmd === 'function') {
        try {
          cmd();
        } catch (e) {
          catalyst.log('Error executing command:', e);
        }
      }
    }
  };

  // Override push to execute commands immediately
  catalyst.cmd.push = function(cmd) {
    if (typeof cmd === 'function') {
      try {
        cmd();
      } catch (e) {
        catalyst.log('Error executing command:', e);
      }
    }
    return Array.prototype.push.call(this, cmd);
  };

  // Process existing commands
  catalyst._processCommandQueue();

  catalyst.log('Catalyst SDK v1.0.0 loaded');

})(window);
