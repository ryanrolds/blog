"use strict";

process.on("unhandledException", (err) => {
  console.error(err);
  process.exist(1);
});

process.on("unhandledRejection", (err) => {
  console.error(err);
  process.exit(1);
});

const express = require("express");
const config = require("./src/config");
const db = require("./src/db");
const posts = require("./src/posts");
const Bluebird = require("bluebird");

const app = express();
// Error handling
app.on("error", (err) => {
  throw err;
});

// Views
app.set("view engine", "pug");

// Middleware


// Routes
app.get("/", (req, res, next) => {
  res.render("index");
});

app.get("/post/:id", (req, res, next) => {
  if (!req.params.id) {
    return res.render("404");
  }

  posts.get(req.params.id).then((post) => {
    if (!post) {
      return res.render("404");
    }

    res.render("post", {"post": post});
  }).catch((err) => {
    res.render("500", {"err": err});
  });
});

// Setup database connection
db.connect().then((client) => {
  return client.query("SELECT count(*) FROM posts").then((result) => {
    client.release();
    return result;
  });
}).then((result) => {
  return new Bluebird((resolve, reject) => {
    app.listen(config.port, () => {
      console.log("Started on %d", config.port);
      return resolve();
    });
  });
}).catch((err) => {
  console.error("Problem starting service", err);
  process.exit(1);
});
