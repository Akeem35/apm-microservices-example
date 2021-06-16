'use strict';
const nr = require('newrelic')
const cache = require('memory-cache');

const OPERATION_CREATE = 'CREATE',
      OPERATION_DELETE = 'DELETE';

class TodoController {
    constructor({redisClient, logChannel}) {
        this._redisClient = redisClient;
        this._logChannel = logChannel;
    }

    // TODO: these methods are not concurrent-safe
    list (req, res) {
        var self = this;
        const test = nr.startSegment('list-items', true, function() {
            const data = self._getTodoData(req.user.username)
            return data.items;
        })
        
        res.json(test)
    }

    create (req, res) {
        // TODO: must be transactional and protected for concurrent access, but
        // the purpose of the whole example app it's enough
        var self = this;

        const result = nr.startSegment('create-item', true, function() {
            const data = self._getTodoData(req.user.username)
            const todo = {
                content: req.body.content,
                id: data.lastInsertedID
            }
            data.items[data.lastInsertedID] = todo
            data.lastInsertedID++
            self._setTodoData(req.user.Fusername, data)
            self._logOperation(OPERATION_CREATE, req.user.username, todo.id)
            return todo;
        });
        
        res.json(result)
    }

    delete (req, res) {

        var self = this;
        const data = this._getTodoData(req.user.username)
        const id = req.params.taskId

        nr.startSegment('delete-item', true, function() {
            delete data.items[id]
            self._setTodoData(req.user.username, data)
        });

        this._logOperation(OPERATION_DELETE, req.user.username, id)
        res.status(204)
        res.send()
    }

    _logOperation(opName, username, todoId) {
        var self = this;
        nr.startSegment('logging-operation', true, function() {
            self._redisClient.publish(
                self._logChannel,
                JSON.stringify({
                    opName,
                    username,
                    todoId,
                }),
                function(err) {
                    if (err) {}
                }
            )
        }); 
    }

    _getTodoData (userID) {
        var self = this;

        var result = nr.startSegment('getting-items', true, function() {
            var data = cache.get(userID)
            if (data == null) {
                data = {
                    items: {
                        '1': {
                            id: 1,
                            content: "Create new todo",
                        },
                        '2': {
                            id: 2,
                            content: "Update me",
                        },
                        '3': {
                            id: 3,
                            content: "Delete example ones",
                        }
                    },
                    lastInsertedID: 3
                }

                self._setTodoData(userID, data)
                self._logOperation('GET', userID, data)
            }

            return data
        });
        
        return result
    }

    _setTodoData (userID, data) {
        nr.startSegment('setting-items', true, function() {
            cache.put(userID, data)
        });

        this._logOperation('SET', userID, data)
    }
}

module.exports = TodoController

