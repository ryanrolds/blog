"use strict";

if (!process.env.PG_URL) {
  throw new Error("Envar PG_URL must be set");
}

exports.port = process.env.PORT || 8080;
exports.pg_url = process.env.PG_URL;
