const mongoose = require('mongoose');
require("dotenv").config();
async function dbConnect(){
    mongoose.connect(process.env.DB_URL).then(()=>{
        console.log("Connected to MongoDB successfully!!!");
    }).catch(err=>{
        console.log("Connection Error: "+err);
    });
}
module.exports = dbConnect;