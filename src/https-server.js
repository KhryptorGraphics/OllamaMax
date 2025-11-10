/**
 * HTTPS Server Configuration
 * Wraps the Express app with HTTPS support
 */

const https = require('https');
const http = require('http');
const fs = require('fs');
const path = require('path');

class HTTPSServer {
  constructor(app) {
    this.app = app;
    this.httpServer = null;
    this.httpsServer = null;
  }

  /**
   * Start server with optional HTTPS
   */
  start(port = 13000) {
    const sslEnabled = process.env.SSL_ENABLED === 'true';

    if (sslEnabled) {
      return this.startHTTPS(port);
    } else {
      return this.startHTTP(port);
    }
  }

  /**
   * Start HTTP server
   */
  startHTTP(port) {
    return new Promise((resolve, reject) => {
      this.httpServer = http.createServer(this.app);

      this.httpServer.listen(port, () => {
        console.log(`🌐 HTTP Server listening on port ${port}`);
        resolve(this.httpServer);
      });

      this.httpServer.on('error', reject);
    });
  }

  /**
   * Start HTTPS server
   */
  startHTTPS(port) {
    return new Promise((resolve, reject) => {
      try {
        // Load SSL certificates
        const certPath = process.env.SSL_CERT_PATH || path.join(__dirname, '../certs/server.crt');
        const keyPath = process.env.SSL_KEY_PATH || path.join(__dirname, '../certs/server.key');
        const dhParamPath = process.env.SSL_DH_PARAM_PATH || path.join(__dirname, '../certs/dhparam.pem');

        // Check if certificate files exist
        if (!fs.existsSync(certPath)) {
          throw new Error(`SSL certificate not found: ${certPath}`);
        }
        if (!fs.existsSync(keyPath)) {
          throw new Error(`SSL key not found: ${keyPath}`);
        }

        const options = {
          cert: fs.readFileSync(certPath),
          key: fs.readFileSync(keyPath)
        };

        // Add DH parameters if available
        if (fs.existsSync(dhParamPath)) {
          options.dhparam = fs.readFileSync(dhParamPath);
          console.log('✓ Using Diffie-Hellman parameters');
        }

        // Security options
        options.honorCipherOrder = true;
        options.ciphers = [
          'ECDHE-ECDSA-AES128-GCM-SHA256',
          'ECDHE-RSA-AES128-GCM-SHA256',
          'ECDHE-ECDSA-AES256-GCM-SHA384',
          'ECDHE-RSA-AES256-GCM-SHA384',
          'ECDHE-ECDSA-CHACHA20-POLY1305',
          'ECDHE-RSA-CHACHA20-POLY1305',
          'DHE-RSA-AES128-GCM-SHA256',
          'DHE-RSA-AES256-GCM-SHA384'
        ].join(':');

        // Create HTTPS server
        this.httpsServer = https.createServer(options, this.app);

        this.httpsServer.listen(port, () => {
          console.log(`🔒 HTTPS Server listening on port ${port}`);
          console.log(`✓ SSL/TLS enabled`);
          resolve(this.httpsServer);
        });

        this.httpsServer.on('error', reject);

        // Optionally start HTTP server for redirect
        if (process.env.HTTP_REDIRECT_ENABLED === 'true') {
          const httpPort = process.env.HTTP_PORT || 80;
          this.startHTTPRedirect(httpPort, port);
        }

      } catch (error) {
        reject(error);
      }
    });
  }

  /**
   * Start HTTP server that redirects to HTTPS
   */
  startHTTPRedirect(httpPort, httpsPort) {
    const redirectApp = require('express')();

    redirectApp.use((req, res) => {
      const host = req.headers.host.split(':')[0];
      const httpsUrl = `https://${host}:${httpsPort}${req.url}`;
      res.redirect(301, httpsUrl);
    });

    this.httpServer = http.createServer(redirectApp);

    this.httpServer.listen(httpPort, () => {
      console.log(`🔀 HTTP redirect server listening on port ${httpPort} → ${httpsPort}`);
    });
  }

  /**
   * Stop all servers
   */
  stop() {
    return new Promise((resolve) => {
      const promises = [];

      if (this.httpServer) {
        promises.push(new Promise((res) => {
          this.httpServer.close(() => {
            console.log('✓ HTTP server stopped');
            res();
          });
        }));
      }

      if (this.httpsServer) {
        promises.push(new Promise((res) => {
          this.httpsServer.close(() => {
            console.log('✓ HTTPS server stopped');
            res();
          });
        }));
      }

      Promise.all(promises).then(resolve);
    });
  }

  /**
   * Get the active server instance
   */
  getServer() {
    return this.httpsServer || this.httpServer;
  }
}

module.exports = HTTPSServer;

