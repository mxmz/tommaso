const PROXY_CONFIG = {
    "/login/": {
      "target": "http://localhost:7997",
      "secure": false,
      "bypass": function (req, res, proxyOptions) {
        req.headers["X-Forwarded-For"] = "9.9.12.12, 42.12.2.11";
      }
    },
    "/api/": {
      "target": "http://localhost:7997",
      "secure": false
    },
    "/.well-known/": {
      "target": "http://localhost:7997",
      "secure": false
    }
  }
  
  module.exports = PROXY_CONFIG;
  