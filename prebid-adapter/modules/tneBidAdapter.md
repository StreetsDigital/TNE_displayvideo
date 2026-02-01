# Overview

```
Module Name: TNE Bid Adapter
Module Type: Bidder Adapter
Maintainer: engineering@thenexusengine.com
```

# Description

Connects to the TNE Catalyst exchange for header bidding via OpenRTB 2.5.
Supports banner and video (instream/outstream) media types.

# Bid Params

| Name | Scope | Description | Example | Type |
|------|-------|-------------|---------|------|
| `publisherId` | required | Publisher ID assigned by TNE | `'pub-123'` | `string` |
| `placementId` | optional | Placement identifier for reporting granularity | `'sidebar-300x250'` | `string` |
| `bidFloor` | optional | Minimum CPM bid floor in USD | `0.50` | `number` |
| `bidFloorCur` | optional | Currency for bidFloor (defaults to `'USD'`) | `'USD'` | `string` |
| `endpoint` | optional | Override exchange base URL (for self-hosted instances via `tneCatalyst` alias) | `'https://exchange.yourdomain.com'` | `string` |
| `custom` | optional | Custom key-value targeting passed to the exchange | `{ "section": "sports" }` | `object` |

# Banner Ad Unit Example

```javascript
var adUnits = [
  {
    code: 'banner-div',
    mediaTypes: {
      banner: {
        sizes: [
          [300, 250],
          [728, 90],
        ],
      },
    },
    bids: [
      {
        bidder: 'tne',
        params: {
          publisherId: 'pub-123',
        },
      },
    ],
  },
];
```

# Video Ad Unit Example (Instream)

```javascript
var adUnits = [
  {
    code: 'video-div',
    mediaTypes: {
      video: {
        context: 'instream',
        playerSize: [640, 480],
        mimes: ['video/mp4', 'video/webm'],
        protocols: [2, 3, 5, 6],
        minduration: 5,
        maxduration: 30,
      },
    },
    bids: [
      {
        bidder: 'tne',
        params: {
          publisherId: 'pub-123',
          placementId: 'video-hero',
        },
      },
    ],
  },
];
```

# Video Ad Unit Example (Outstream)

```javascript
var adUnits = [
  {
    code: 'outstream-div',
    mediaTypes: {
      video: {
        context: 'outstream',
        playerSize: [640, 480],
        mimes: ['video/mp4'],
        protocols: [2, 3, 5, 6],
      },
    },
    bids: [
      {
        bidder: 'tne',
        params: {
          publisherId: 'pub-123',
          placementId: 'outstream-article',
        },
      },
    ],
  },
];
```

# Multi-Format Ad Unit Example

```javascript
var adUnits = [
  {
    code: 'multi-div',
    mediaTypes: {
      banner: {
        sizes: [[300, 250]],
      },
      video: {
        context: 'outstream',
        playerSize: [300, 250],
        mimes: ['video/mp4'],
        protocols: [2, 3],
      },
    },
    bids: [
      {
        bidder: 'tne',
        params: {
          publisherId: 'pub-123',
          placementId: 'multi-sidebar',
          bidFloor: 1.5,
        },
      },
    ],
  },
];
```

# Self-Hosted / Alias Example (`tneCatalyst`)

Operators running their own TNE Catalyst exchange instance use the `tneCatalyst` alias
with a custom `endpoint` pointing to their server:

```javascript
var adUnits = [
  {
    code: 'banner-div',
    mediaTypes: {
      banner: {
        sizes: [[300, 250], [728, 90]],
      },
    },
    bids: [
      {
        bidder: 'tneCatalyst',
        params: {
          publisherId: 'pub-456',
          endpoint: 'https://exchange.yourdomain.com',
        },
      },
    ],
  },
];
```
