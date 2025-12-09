const http = require('http');

const port = process.env.PORT || 3001;

let counter = 0;

const server = http.createServer((req, res) => {
  if (req.method === 'GET' && req.url === '/pingpong') {
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/plain');
    res.end(`pong ${counter++}`);
  } else if (req.method === 'GET' && req.url === '/pings') {
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/plain');
    res.end(counter.toString());
  } else {
    res.statusCode = 404;
    res.setHeader('Content-Type', 'text/plain');
    res.end('Not found');
  }
});

server.listen(port, () => {
  console.log(`HTTP server listening on port ${port}`);
});