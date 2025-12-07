const fs = require('fs');
const path = require('path');

const filePath = path.join('/usr/src/app/shared', 'status.txt');

const generatedString = cryptoRandomString();

function cryptoRandomString() {
  return require('crypto').randomUUID();
}

function writeLine() {
  const timestamp = new Date().toISOString();
  const line = `${timestamp}: ${generatedString}\n`;

  // append to the file in the shared volume
  fs.appendFile(filePath, line, (err) => {
    if (err) {
      console.error('Failed to write to file:', err);
    }
  });

  setTimeout(writeLine, 5000);
}

writeLine();
