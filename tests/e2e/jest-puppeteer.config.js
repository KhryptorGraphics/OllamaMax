module.exports = {
  launch: {
    headless: true,
    defaultViewport: null,
    args: ['--no-sandbox', '--disable-setuid-sandbox']
  },
  browserContext: 'default',
  server: {
    command: 'echo "Using external server (nginx in compose)"',
    port: 80,
    launchTimeout: 60000,
    debug: false
  }
};

