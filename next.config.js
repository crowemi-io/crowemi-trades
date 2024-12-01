// next.config.js

module.exports = {
    async headers() {
      return [
        {
          source: '/(.*)', // Apply these headers to all routes
          headers: [
            {
              key: 'Content-Security-Policy',
              value: `
                default-src 'self';
                script-src 'self' 'unsafe-eval' 'unsafe-inline';
                style-src 'self' 'unsafe-inline';
                img-src 'self' data:;
                connect-src 'self' ${process.env.URL};
                font-src 'self';
                object-src 'none';
                frame-ancestors 'none';
                base-uri 'self';
                form-action 'self';
              `.replace(/\s{2,}/g, ' ').trim()
            }
          ]
        }
      ]
    }
  }