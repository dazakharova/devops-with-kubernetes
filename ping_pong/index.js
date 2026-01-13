const http = require('http');
const { Pool } = require('pg');

const port = process.env.PORT || 3001;
const connectionString = process.env.DATABASE_URL;

if (!connectionString) {
  console.error('DATABASE_URL must be set');
  process.exit(1);
}

const pool = new Pool({ connectionString });

async function dbReady() {
    await pool.query('SELECT 1');
}

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
    try {
        if (req.method === 'GET' && req.url === '/readyz') {
            await pool.query('SELECT 1');
            res.statusCode = 200;
            res.setHeader('Content-Type', 'text/plain');
            return res.end('ready\n');
        }

        if (req.method === 'GET' && req.url === '/') {
            const value = await incrementCounter();
            res.statusCode = 200;
            res.setHeader('Content-Type', 'text/plain');
            return res.end(`pong ${value}\n`);
        }

        if (req.method === 'GET' && req.url === '/pings') {
            const value = await getCounter();
            res.statusCode = 200;
            res.setHeader('Content-Type', 'text/plain');
            return res.end(value.toString());
        }

        res.statusCode = 404;
        res.end('Not found');
    } catch (err) {
        console.error(err);
        res.statusCode = 503; // critical: don't crash the process
        res.setHeader('Content-Type', 'text/plain');
        res.end('db not ready\n');
    }
});

server.listen(port, () => {
  console.log(`HTTP server listening on port ${port}`);
});