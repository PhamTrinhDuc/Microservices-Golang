const fs = require('fs');
const path = require('path');

const filePath = path.join(__dirname, 'products', 'ProductsRouter.jsx');
const content = fs.readFileSync(filePath, 'utf8');
const lines = content.split('\n');

const query = /router\.(get|post|put|delete|patch)\(['"]\/[^'"]*['"]/gi;
console.log("Matching routes:");
lines.forEach((line, index) => {
    if (query.test(line)) {
        console.log(`${index + 1}: ${line.trim()}`);
    }
});
