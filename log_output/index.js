const http = require('http');

const generatedString = cryptoRandomString()
let lastTimestamp = null;

function cryptoRandomString() {
  return require('crypto').randomUUID()
}

function printString() {
  const timestamp = new Date().toISOString()
  lastTimestamp = timestamp;
  console.log(`${timestamp}: ${generatedString}`)
  setTimeout(printString, 5000)
}

// Start repeated logging
printString()

const port = process.env.PORT || 3000;

const server = http.createServer((req, res) => {
  if (req.method === 'GET' && req.url === '/status') {
    const currentTimestamp = lastTimestamp || new Date().toISOString();

    res.statusCode = 200;
    res.setHeader('Content-Type', 'application/json');
    res.end(
        JSON.stringify({
          timestamp: currentTimestamp,
          random: generatedString,
        })
    );
  } else {
    res.statusCode = 404;
    res.setHeader('Content-Type', 'text/plain');
    res.end('Not found');
  }
});

server.listen(port, () => {
  console.log(`HTTP server listening on port ${port}`);
});