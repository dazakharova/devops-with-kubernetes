const http = require('http');
const fs = require('fs');
const path = require('path');

const port = process.env.PORT || 3000;
const statusFile = path.join('/usr/src/app/shared', 'status.txt');
const counterFile = path.join('/usr/src/app/data', 'counter.txt');

function readFileSafe(filePath, defaultValue, cb) {
  fs.readFile(filePath, 'utf8', (err, data) => {
    if (err) {
      if (err.code === 'ENOENT') {
        return cb(null, defaultValue);
      }
      return cb(err);
    }
    cb(null, data);
  });
}

const server = http.createServer((req, res) => {
  if (req.method === 'GET' && req.url === '/status') {
    readFileSafe(statusFile, '', (statusErr, statusData) => {
      if (statusErr) {
        console.error('Error reading status file:', statusErr);
        res.statusCode = 500;
        res.setHeader('Content-Type', 'text/plain');
        return res.end('Error reading status file');
      }

      readFileSafe(counterFile, '0', (counterErr, counterData) => {
        if (counterErr) {
          console.error('Error reading counter file:', counterErr);
          res.statusCode = 500;
          res.setHeader('Content-Type', 'text/plain');
          return res.end('Error reading counter file');
        }

        const trimmedStatus = statusData.trimEnd();
        const trimmedCounter = counterData.toString().trim() || '0';

        const body =
            (trimmedStatus ? trimmedStatus + '\n' : '') +
            `ping-pong count: ${trimmedCounter}\n`;

        res.statusCode = 200;
        res.setHeader('Content-Type', 'text/plain');
        res.end(body);
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
