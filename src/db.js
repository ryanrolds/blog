"use strict";

const config = require("./config");
const Pool = require("pg").Pool;
const pool = new Pool({"connectionString": config.pg_url});

pool.on("error", (err) => {
  console.error("Unexpected pool error", err);
  process.exit(1);
});

module.exports = pool;
