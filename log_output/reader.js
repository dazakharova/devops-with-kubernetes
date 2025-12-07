const http = require('http');
const fs = require('fs');
const path = require('path');

const port = process.env.PORT || 3000;
const filePath = path.join('/usr/src/app/shared', 'status.txt');

const server = http.createServer((req, res) => {
  if (req.method === 'GET' && req.url === '/status') {
    fs.readFile(filePath, 'utf8', (err, data) => {
      if (err) {
        res.statusCode = 500;
        res.setHeader('Content-Type', 'text/plain');
        res.end('Error reading status file');
        return;
      }

      res.statusCode = 200;
      res.setHeader('Content-Type', 'text/plain');
      res.end(data);
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
