
const dotenv = require('dotenv');
dotenv.config();

const express = require('express');
const { altairExpress } = require('altair-express-middleware');

const server = express();

// Mount your Altair GraphQL client
server.use('/', altairExpress({
  endpointURL: `${process.env.GQL_SERVER_ADDRESS}`,
  initialQuery: `# Insert a query or mutation. You can also consult the docs on the right to see the schema!`,
}));


server.listen(process.env.ALTAIR_PORT, ()=>{
    console.log('started Altair GraphQL IDE 🎉')
})