const http = require('http');
const fs = require('fs');
const path = require('path');

const port = process.env.PORT || 3000;
const statusFile = path.join('/usr/src/app/shared', 'status.txt');
const infoFile = path.join('/usr/src/app/config', 'information.txt');
const message = process.env.MESSAGE || '';
const pingsUrl = 'http://ping-pong-svc:2345/pings';

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

async function getPings() {
  const res = await fetch(pingsUrl);
  if (!res.ok) {
    throw new Error(`failed to fetch pings: ${res.status} ${res.statusText}`);
  }

  const text = await res.text();
  const n = parseInt(text, 10);

  if (Number.isNaN(n)) {
    throw new Error(`invalid pings value: ${text}`);
  }

  return n;
}

const server = http.createServer((req, res) => {
  if (req.method === 'GET' && req.url === '/') {
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/plain');
    return res.end('ok\n');
  } else if (req.method === 'GET' && req.url === '/readyz') {
    getPings()
        .then(() => {
            res.statusCode = 200;
            res.setHeader('Content-Type', 'text/plain');
            res.end('ready\n');
        })
        .catch(() => {
            res.statusCode = 503;
            res.setHeader('Content-Type', 'text/plain');
            res.end('not ready\n');
        });
  } else if (req.method === 'GET' && req.url === '/status') {
    readFileSafe(statusFile, '', (statusErr, statusData) => {
      if (statusErr) {
        console.error('Error reading status file:', statusErr);
        res.statusCode = 500;
        res.setHeader('Content-Type', 'text/plain');
        return res.end('Error reading status file');
      } 

      fs.readFile(infoFile, 'utf8', (infoErr, infoData) => {
        const fileLine = infoErr
            ? 'file content: <error reading file>'
            : `file content: ${infoData.trim()}`;

        const envLine = `env variable: MESSAGE=${message || '<empty>'}`;

        getPings()
            .then((pings) => {
              const trimmedStatus = statusData.trimEnd();
              const body =
                  fileLine + '\n' +
                  envLine + '\n' +
                  (trimmedStatus ? trimmedStatus + '\n' : '') +
                  `Ping / Pongs: ${pings}\n`;

              res.statusCode = 200;
              res.setHeader('Content-Type', 'text/plain');
              res.end(body);
            })
            .catch((err) => {
              console.error('Error fetching pings:', err);
              res.statusCode = 500;
              res.setHeader('Content-Type', 'text/plain');
              res.end('Error fetching pings...');
            });
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
