'use strict';

exports.up = function(db) {
  return db.createTable('posts', {
    "id": {
      "type": "string",
      "primaryKey": true,
    },
    "views": {
      "type": "int",
      "defaultValue": 0
    }
  });
};

exports.down = function(db) {
  return db.dropTable('posts');
};
