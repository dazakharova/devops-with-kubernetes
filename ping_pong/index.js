const http = require('http');
const { Pool } = require('pg');

const port = process.env.PORT || 3001;
const connectionString = process.env.DATABASE_URL;

if (!connectionString) {
  console.error('DATABASE_URL must be set');
  process.exit(1);
}

const pool = new Pool({ connectionString });

async function getCounter() {
  const result = await pool.query(
      'SELECT value FROM pingpong_counter WHERE id = 1'
  );
  if (result.rowCount === 0) {
    throw new Error('counter row missing');
  }
  return result.rows[0].value;
}

async function incrementCounter() {
  const result = await pool.query(
      'UPDATE pingpong_counter SET value = value + 1 WHERE id = 1 RETURNING value'
  );
  if (result.rowCount === 0) {
    throw new Error('counter row missing');
  }
  return result.rows[0].value;
}

const server = http.createServer(async (req, res) => {
  if (req.method === 'GET' && req.url === '/') {
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/plain');
    return res.end('ok\n');
  } else if (req.method === 'GET' && req.url === '/pingpong') {
    const value = await incrementCounter();
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/plain');
    res.end(`pong ${value}`);
  } else if (req.method === 'GET' && req.url === '/pings') {
    const value = await getCounter();
    res.statusCode = 200;
    res.setHeader('Content-Type', 'text/plain');
    res.end(value.toString());
  } else {
    console.error('Error handling request');
    res.statusCode = 500;
    res.setHeader('Content-Type', 'text/plain');
    res.end('Internal server error');
  }
});

server.listen(port, () => {
  console.log(`HTTP server listening on port ${port}`);
});