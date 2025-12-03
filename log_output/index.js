const generatedString = cryptoRandomString()

function cryptoRandomString() {
  return require('crypto').randomUUID()
}

function printString() {
  const timestamp = new Date().toISOString()
  console.log(`${timestamp}: ${generatedString}`)
  setTimeout(printString, 5000)
}

// Start repeated logging
printString()
