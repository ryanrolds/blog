"use strict";

const Bluebird = require("bluebird");
const lodash = require("lodash");
const db = require("./db");
const fs = require("fs");
const posts = {};

exports.get = function(id) {
  return new Bluebird((resolve, reject) => {
    if (lodash.has(posts, id)) {
      return resolve(lodash.get(posts, id));
    }

    // TODO cleanup id and make sure it's not a path
    fs.readFile("../posts/" + id, "utf8", (err, data) => {
      if (err) {
        return reject(err);
      }

      // TODO parse markdown in to something useful
      // TODO secure this by removing periods and things
      lodash.set(posts, id);
      return resolve(data);
    });
  }).then((content) => {
    return db.connect().then((client) => {
      // Check for existing post
      let query = "SELECT * FROM posts WHERE id = $1";
      let vals = [id];
      return client.query(query, vals).then((result) => {
        if (!result.rows.length) {
          // Create post if it does not exist already
          return createPage(client, id).then((result) => {
            return result.rows[0];
          });
        }

        return result.rows[0];
      }).then((post) => {
        // Increment view count
        return addPageView(client, id).then(() => {
          client.release();
          return post;
        });
      }).catch((err) => {
        client.release();
        throw err;
      });
    }).then((post) => {
      // Take markdown file content and add to post
      post.content = content;
      return post;
    });
  });
};

function addPageView(client, id) {
  let query = "UPDATE posts SET views = views + 1 WHERE id = $1";
  let vals = [id];
  return client.query(query, vals);
}

function createPage(client, id) {
  let query = "INSERT INTO posts (id) VALUES ($1)";
  let vals = [id];
  return client.query(query, vals);
}
