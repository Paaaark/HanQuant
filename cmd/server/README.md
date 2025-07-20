# HanQuant Server API Documentation

This document describes all REST API endpoints provided by the HanQuant server. Each endpoint includes required/optional parameters, a summary, response formats, status codes, and possible error messages.

---

<details>
<summary><strong>POST /auth/register</strong> — Register a new user</summary>

**Summary:**
Registers a new user with a username and password.

**Request Body:**

```json
{
  "username": "string (3-50 chars, required)",
  "password": "string (min 6 chars, required)"
}
```

**Response:**

- `201 Created` — Registration successful, returns `{}`
- `400 Bad Request` — Invalid JSON, missing fields, or invalid format
- `409 Conflict` — Username already exists
- `500 Internal Server Error` — Server error

**Example Error Responses:**

```json
{"error": "username and password are required"}
{"error": "username must be between 3 and 50 characters"}
{"error": "password must be at least 6 characters long"}
{"error": "username already exists"}
```

</details>

<details>
<summary><strong>POST /auth/login</strong> — User login</summary>

**Summary:**
Authenticates a user and returns a JWT and refresh token.

**Request Body:**

```json
{
  "username": "string (required)",
  "password": "string (required)"
}
```

**Response:**

- `200 OK` — Success

```json
{
  "token": "JWT string",
  "refresh_token": "string",
  "expires_in": 3600,
  "refresh_expires_at": "RFC3339 timestamp"
}
```
Where:
- `token` — JWT to include in the `Authorization` header for authenticated
  requests.
- `refresh_token` — long‑lived token used with `/auth/refresh` to obtain a new
  JWT.
- `expires_in` — seconds until the JWT expires.
- `refresh_expires_at` — RFC3339 timestamp when the refresh token becomes
  invalid.

- `400 Bad Request` — Invalid JSON or missing fields
- `401 Unauthorized` — Invalid username or password
- `500 Internal Server Error` — Server error

**Example Error Responses:**

```json
{"error": "username and password are required"}
{"error": "invalid username or password"}
```

</details>

<details>
<summary><strong>POST /auth/refresh</strong> — Refresh JWT</summary>

**Summary:**
Refreshes the JWT using a valid refresh token.

**Request Body:**

```json
{
  "refresh_token": "string (required)"
}
```

**Response:**

- `200 OK` — Success

```json
{
  "token": "JWT string",
  "expires_in": 3600
}
```
Where:
- `token` — New JWT for subsequent requests.
- `expires_in` — Seconds until the returned JWT expires.

- `400 Bad Request` — Invalid request
- `401 Unauthorized` — Invalid or expired refresh token
- `500 Internal Server Error` — Server error

**Example Error Responses:**

```json
{"error": "invalid or expired refresh token"}
{"error": "user not found"}
```

</details>

<details>
<summary><strong>POST /accounts</strong> — Link a bank account (KIS)</summary>

**Summary:**
Links a KIS account and API keys to the authenticated user.

**Headers:**

- `Authorization: Bearer <JWT>`

**Request Body:**

```json
{
  "account_id": "string (required)",
  "app_key": "string (required)",
  "app_secret": "string (required)",
  "cano": "string (required)",
  "is_mock": "boolean (optional)"
}
```

**Response:**

- `201 Created` — Account linked, returns the account object
- `400 Bad Request` — Invalid request or missing fields
- `409 Conflict` — Account already linked
- `500 Internal Server Error` — Server error

The returned object has the following fields:
- `id` — Numeric identifier for the link record.
- `user_id` — ID of the owner user.
- `account_id` — Friendly identifier for the account.
- `enc_cano` — Encrypted CANO value (8‑2 account number).
- `enc_app_key` — Encrypted API key.
- `enc_app_secret` — Encrypted API secret.
- `is_mock` — `true` when using a mock/paper account.
- `created_at` — Timestamp the link was created.

**Example Error Responses:**

```json
{"error": {"code": "VALIDATION", "message": "missing required fields"}}
{"error": {"code": "CONFLICT", "message": "account already linked"}}
```

</details>

<details>
<summary><strong>GET /accounts</strong> — List linked accounts</summary>

**Summary:**
Returns all KIS accounts linked to the authenticated user.

**Headers:**

- `Authorization: Bearer <JWT>`

**Response:**

- `200 OK` — Array of account objects
- `500 Internal Server Error` — Server error

Each object in the array contains the same fields returned by
`POST /accounts` (`id`, `user_id`, `account_id`, `enc_cano`,
`enc_app_key`, `enc_app_secret`, `is_mock`, `created_at`).

**Example Error Response:**

```json
{ "error": { "code": "DB", "message": "..." } }
```

</details>

<details>
<summary><strong>DELETE /accounts/{id}</strong> — Unlink a bank account</summary>

**Summary:**
Unlinks a KIS account from the authenticated user.

**Headers:**

- `Authorization: Bearer <JWT>`

**Path Parameter:**

- `id` — Account ID (integer)

**Response:**

- `204 No Content` — Account unlinked
- `400 Bad Request` — Invalid account ID
- `500 Internal Server Error` — Server error

**Example Error Responses:**

```json
{"error": {"code": "VALIDATION", "message": "invalid account id"}}
{"error": {"code": "DB", "message": "..."}}
```

</details>

<details>
<summary><strong>GET /portfolio?account_id=...</strong> — Get portfolio for a linked account</summary>

**Summary:**
Returns the portfolio for a specific linked account.

**Headers:**

- `Authorization: Bearer <JWT>`

**Query Parameter:**

- `account_id` — The linked account ID (string, required)

**Response:**

- `200 OK` — Portfolio data
- `400 Bad Request` — Missing account_id
- `404 Not Found` — No linked account for user
- `500 Internal Server Error` — Server or KIS error

Successful responses look like:

```json
{
  "as_of": "RFC3339 timestamp",
  "positions": [
    {
      "Symbol": "string",
      "Name": "string",
      "TradeType": "string",
      "HoldingQty": "string",
      "OrderableQty": "string",
      "AvgPrice": "string",
      "PurchaseAmount": "string",
      "CurrentPrice": "string",
      "EvaluationAmount": "string",
      "UnrealizedPnl": "string",
      "UnrealizedPnlRate": "string",
      "FluctuationRate": "string"
    }
  ],
  "summary": {
    "TotalDeposit": "string",
    "D2Deposit": "string",
    "TotalPurchaseAmount": "string",
    "TotalEvaluationAmount": "string",
    "TotalUnrealizedPnl": "string",
    "NetAsset": "string",
    "AssetChangeAmount": "string",
    "AssetChangeRate": "string"
  }
}
```

**Example Error Responses:**

```json
{"error": {"code": "VALIDATION", "message": "account_id required"}}
{"error": {"code": "ACCOUNT_NOT_FOUND", "message": "No linked account for user"}}
```

</details>

<details>
<summary><strong>POST /orders</strong> — Place an order</summary>

**Summary:**
Places a buy or sell order for a stock using a linked account.

**Headers:**

- `Authorization: Bearer <JWT>`

**Request Body:**

```json
{
  "account_id": "string (required)",
  "symbol": "string (required)",
  "side": "buy|sell (required)",
  "qty": "number (required)",
  "order_type": "string (required)",
  "limit_price": "number (optional)"
}
```

**Response:**

- `200 OK` — Order placed, returns order object
- `400 Bad Request` — Invalid request or missing fields
- `404 Not Found` — No linked account for user
- `500 Internal Server Error` — Server or KIS error

Example order object:

```json
{
  "id": 1,
  "user_account_id": 2,
  "symbol": "005930",
  "side": "BUY",
  "qty": 10.0,
  "order_type": "LIMIT",
  "limit_price": 70000,
  "status": "PENDING",
  "kis_order_id": "...",
  "created_at": "RFC3339 timestamp"
}
```

**Example Error Responses:**

```json
{"error": {"code": "VALIDATION", "message": "missing or invalid fields"}}
{"error": {"code": "ACCOUNT_NOT_FOUND", "message": "No linked account for user"}}
```

</details>

<details>
<summary><strong>GET /orders/{id}</strong> — Get order details</summary>

**Summary:**
Returns details for a specific order.

**Headers:**

- `Authorization: Bearer <JWT>`

**Path Parameter:**

- `id` — Order ID (integer)

**Response:**

- `200 OK` — Order object
- `400 Bad Request` — Invalid order ID
- `404 Not Found` — Order not found
- `500 Internal Server Error` — Server error

The returned object shares the same shape as in `POST /orders`. See above for a
field-by-field example of the order JSON structure.

**Example Error Responses:**

```json
{"error": {"code": "VALIDATION", "message": "invalid order id"}}
{"error": {"code": "NOT_FOUND", "message": "order not found"}}
```

</details>

<details>
<summary><strong>GET /prices/recent/{symbol}</strong> — Get recent price for a stock</summary>

**Summary:**
Returns the most recent price for a given stock symbol.

**Path Parameter:**

- `symbol` — Stock symbol (string)

**Response:**

- `200 OK` — Price data
- `400 Bad Request` — Invalid path
- `500 Internal Server Error` — Server error

Returns an array of objects:

```json
[
  {
    "Date": "YYYYMMDD",
    "Open": "string",
    "High": "string",
    "Low": "string",
    "Close": "string",
    "Volume": "string",
    "Duration": "D"
  }
]
```
</details>

<details>
<summary><strong>GET /prices/historical/{symbol}?from=...&to=...&duration=...</strong> — Get historical prices</summary>

**Summary:**
Returns historical price data for a stock symbol.

**Query Parameters:**

- `from` — Start date (string, required)
- `to` — End date (string, required)
- `duration` — Duration (string, required)

**Response:**

- `200 OK` — Historical price data
- `400 Bad Request` — Missing query parameters
- `500 Internal Server Error` — Server error

Response body is the same array of objects described for
`GET /prices/recent/{symbol}` with the requested date range.
</details>

<details>
<summary><strong>GET /ranking/fluctuation</strong> — Top fluctuation stocks</summary>

**Summary:**
Returns stocks with the highest price fluctuations.

**Response:**

- `200 OK` — List of stocks
- `500 Internal Server Error` — Server error
Fields are the same as the other ranking endpoints (`Code`, `Name`, `Price`,
etc.).
Each element uses the same fields as `/ranking/fluctuation`.

Each entry has the following properties:
`Code`, `Name`, `Price`, `Change`, `ChangeSign`, `ChangeRate`, `Volume`,
`MarketCap`, and `Rank`.
</details>

<details>
<summary><strong>GET /ranking/volume</strong> — Most traded stocks</summary>

**Summary:**
Returns stocks with the highest trading volume.

**Response:**

- `200 OK` — List of stocks
- `500 Internal Server Error` — Server error
</details>

<details>
<summary><strong>GET /ranking/market-cap</strong> — Top market cap stocks</summary>

**Summary:**
Returns stocks with the highest market capitalization.

**Response:**

- `200 OK` — List of stocks
- `500 Internal Server Error` — Server error
</details>

<details>
<summary><strong>GET /snapshot/multstock?tickers=...</strong> — Get multiple stock snapshots</summary>

**Summary:**
Returns snapshot data for up to 30 stock tickers.

**Query Parameter:**

- `tickers` — Comma-separated list of stock symbols (max 30)

**Response:**

- `200 OK` — Snapshot data
- `400 Bad Request` — Missing tickers or too many tickers
- `500 Internal Server Error` — Server error

Each snapshot entry includes pricing information and order book depth with
fields such as `Code`, `Name`, `Price`, `Change`, `Open`, `High`, `Low`,
`Volume`, `AskPrice`, `BidPrice`, and many more (see `StockSnapshot` struct).
</details>

<details>
<summary><strong>GET /index/{code}</strong> — Get index price</summary>

**Summary:**
Returns the price for a given index code.

**Path Parameter:**

- `code` — Index code (string)

**Response:**

- `200 OK` — Index price data
- `400 Bad Request` — Missing index code
- `500 Internal Server Error` — Server error

Example response:

```json
{
  "IndexCode": "0001",
  "IndexName": "Kospi",
  "Date": "YYYYMMDD",
  "CurrentPrice": "string",
  "ChangeFromPrev": "string",
  "ChangeSign": "string",
  "ChangeRate": "string",
  "Open": "string",
  "High": "string",
  "Low": "string",
  "Volume": "string",
  "RisingCnt": "string",
  "UpperLimitCnt": "string",
  "FlatCnt": "string",
  "FallingCnt": "string",
  "LowerLimitCnt": "string"
}
```
</details>

<details>
<summary><strong>WebSocket: /ws/stocks</strong> — Real-time stock updates</summary>

**Summary:**
Establishes a WebSocket connection for real-time stock updates.

**Response:**

- Real-time JSON messages for subscribed tickers

Clients send messages of the form:

```json
{ "type": "subscribe", "tickers": ["005930", "000660"] }
```

or

```json
{ "type": "unsubscribe", "tickers": ["005930"] }
```

Incoming updates have:

```json
{
  "type": "snapshot",
  "data": [ { /* StockSnapshot fields */ } ]
}
```

If an error occurs, the server sends:

```json
{ "type": "error", "error": "message" }
```
</details>

---

For more details on request/response formats, see the handler code in `internal/handler/` and service logic in `internal/service/`.
