const http = require('http');
const fs = require('fs');
const path = require('path');

const port = process.env.PORT || 3001;

const dataDir = '/usr/src/app/data';
const counterFile = path.join(dataDir, 'counter.txt');

function readCounter(callback) {
  fs.readFile(counterFile, 'utf8', (err, data) => {
    if (err) {
      if (err.code === 'ENOENT') {
        return callback(null, 0);
      }
      return callback(err);
    }

    const parsed = parseInt(data, 10);
    if (isNaN(parsed)) {
      return callback(null, 0);
    }
    callback(null, parsed);
  });
}

function writeCounter(value, callback) {
  fs.writeFile(counterFile, String(value), 'utf8', callback);
}

const server = http.createServer((req, res) => {
  if (req.method === 'GET' && req.url === '/pingpong') {
    readCounter((readErr, current) => {
      if (readErr) {
        console.error('Error reading counter:', readErr);
        res.statusCode = 500;
        res.setHeader('Content-Type', 'text/plain');
        return res.end('Internal server error');
      }

      const next = current + 1;

      writeCounter(next, (writeErr) => {
        if (writeErr) {
          console.error('Error writing counter:', writeErr);
          res.statusCode = 500;
          res.setHeader('Content-Type', 'text/plain');
          return res.end('Internal server error');
        }

        res.statusCode = 200;
        res.setHeader('Content-Type', 'text/plain');
        res.end(`pong ${next}`);
      });
    });
  } else {
    res.statusCode = 404;
    res.setHeader('Content-Type', 'text/plain');
    res.end('Not found');
  }
});

server.listen(port, () => {
  console.log(`HTTP server listening on port ${port}`);
});