"use strict";

const express = require("express");
const config = require("./src/config");

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
  res.send("Hello World!");
});

app.get("/post/:id", (req, res, next) => {
  if (req.params.id) {
    res.render("404");
  }

  // Lookup blog post
  // Render blog post

  res.send("Post " + req.params.id);
});

app.listen(config.port, () => {
  console.log("Started on %d", config.port);
});
