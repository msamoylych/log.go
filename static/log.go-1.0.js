$(function () {
    const STREAM = 'stream',
        NODE = 'node';

    let Color = (function () {
        const MAX = 20;

        let _idx = 1,
            _store = {},
            _next = function () {
                if (_idx > MAX) {
                    _idx = 1;
                }
                return 'color' + _idx++
            };

        return {
            of: function (type, name) {
                let typeStore = _store[type];
                if (!typeStore) {
                    typeStore = {};
                    _store[type] = typeStore;
                }
                let color = typeStore[name];
                if (!color) {
                    color = _next();
                    typeStore[name] = color;
                }
                return color;
            }
        }
    })();

    let Key = (function () {
        return {
            of: function (node, stream) {
                return node + ':' + stream;
            }
        };
    })();

    let Message = (function () {

        function Message(msg) {
            this._node = msg.node;
            this._stream = msg.stream;
            this.text = msg.text;
            this.key = Key.of(this._node, this._stream);
        }

        Message.prototype.render = function () {
            let msg = $('<div>', {class: 'message'});
            msg.append($('<span>', {class: 'stream', text: this._stream}).addClass(Color.of(STREAM, this._stream)));
            msg.append(' ');
            msg.append($('<span>', {class: 'node', text: this._node}).addClass(Color.of(NODE, this._node)));
            msg.append(' ');
            msg.append($('<span>', {class: 'text', text: this.text}));
            return msg;
        };

        return Message
    })();

    let Screen = (function () {

        function Screen(sources, def) {
            this._sources = sources ? sources : {};
            this._def = !!def;
            this._messages = [];
            this._scroll = true;
        }

        Screen.prototype.render = function () {
            let _this = this;
            let screen = $('<div>', {class: 'screen border rounded'});
            let scrollTop = 0;
            let messages = $('<div>', {class: 'messages'});
            messages.scroll(function () {
                let current = messages.scrollTop();
                if (_this._scroll && current < scrollTop) {
                    _this._scroll = false
                } else if (!_this._scroll && messages.prop('scrollHeight') - messages.prop('clientHeight') === current) {
                    _this._scroll = true
                }
                scrollTop = current;
            });
            screen.append(messages);
            this._msgs = messages;
            let controls = $('<div>', {class: 'screen-controls', style: 'display: none;'});
            let filter = $('<input>', {type: 'text', class: 'form-control', placeholder: 'Filter...'});
            filter.on('input', function () {
                let value = this.value;
                _this._filter = value;
                for (let message of _this._messages) {
                    message.message.toggle(message.text.includes(value))
                }
            });
            controls.append(filter);
            let clear = $('<button>', {type: 'button', class: 'btn btn-secondary', text: 'clear'});
            clear.click(function () {
                messages.empty();
                _this._messages = [];
            });
            controls.append(clear);
            if (!this._def) {
                let close = $('<button>', {type: 'button', class: 'btn btn-secondary', text: 'close'});
                close.click(function () {
                    _this._screen.remove();
                    Controls.delScreen(_this);
                    Screens.delScreen(_this);
                });
                controls.append(close);
            }
            screen.append(controls);
            screen.hover(function () {
                controls.show()
            }, function () {
                controls.hide()
            });
            this._screen = screen;
            return screen;
        };

        Screen.prototype.newMessage = function (msg) {
            if (this.enabled(msg.key)) {
                let message = msg.render();
                if (this._filter && !msg.text.includes(this._filter)) {
                    message.hide();
                }
                msg.message = message;
                this._messages.push(msg);
                this._msgs.append(message);
                if (this._scroll) {
                    this._msgs.scrollTop(this._msgs.prop('scrollHeight') - this._msgs.prop('clientHeight'))
                }
            }
        };

        Screen.prototype.toggle = function (key, enabled) {
            this._sources[key] = enabled;
            this.onToggle();
        };

        Screen.prototype.enabled = function (key) {
            return this._sources[key] || (this._def && this._sources[key] === undefined)
        };

        Screen.prototype.toJSON = function () {
            return this._sources;
        };

        return Screen;
    })();

    let Screens = (function () {
        const SCREENS_KEY = 'screens',
            MAX = 4;
        let _screens = [],
            _messagesCount = 0,
            _load = function () {
                return localStorage.getItem(SCREENS_KEY);
            },
            _save = function () {
                localStorage.setItem(SCREENS_KEY, JSON.stringify(_screens));
            },
            _newScreen = function (sources, def) {
                let screen = new Screen(sources, def);
                screen.onToggle = _save;
                _screens.push(screen);
                $('#screens').append(screen.render());
                return screen
            };

        return {
            init: function () {
                let screenCache = _load();
                let sources = screenCache ? JSON.parse(screenCache) : [{}];
                for (let src of sources) {
                    _newScreen(src, _screens.length === 0)
                }
            },
            handler: function (msg) {
                _messagesCount++;
                let message = new Message(msg);
                for (let screen of _screens) {
                    screen.newMessage(message)
                }
            },
            addScreen: function () {
                if (_screens.length < MAX) {
                    let screen = _newScreen();
                    _save();
                    return screen
                } else {
                    return undefined
                }
            },
            delScreen: function (screen) {
                let idx = _screens.indexOf(screen);
                _screens.splice(idx, 1);
                _save();
            },
            messagesCount: function () {
                return _messagesCount;
            },
            [Symbol.iterator]: function () {
                let _idx = 0;
                return {
                    next: function () {
                        return {
                            value: _screens[_idx],
                            done: _idx++ === _screens.length
                        }
                    }
                }
            }
        };
    })();

    let Control = (function () {

        function Control(type, name, key, parent) {
            this._name = name;
            this._key = key;
            this._parent = parent;
            this._childs = [];
            this._cbxes = new Map();
            this._itemCls = parent ? 'list-group-item-secondary' : 'list-group-item-primary';
            this._color = Color.of(type, name);
            if (parent) {
                parent._childs.push(this);
            }
        }

        Control.prototype.addScreen = function (screen) {
            this._checkbox(screen);
        };

        Control.prototype.delScreen = function (screen) {
            let checkbox = this._cbxes.get(screen);
            this._cbxes.delete(screen);
            checkbox.remove();
        };

        Control.prototype.link = function (control) {
            this._linked = control;
            control._linked = this;
        };

        Control.prototype.render = function () {
            let item = $('<div>', {class: 'list-group-item'}).addClass(this._itemCls);
            this._diode = $('<div>', {class: 'diode'}).addClass(this._color);
            item.append(this._diode);
            item.append($('<div>', {class: 'name', text: this._name}));
            this._checkboxes = $('<div>', {class: 'checkboxes'});
            item.append(this._checkboxes);
            let active = false;
            for (let screen of Screens) {
                active |= this._checkbox(screen);
            }
            if (active) {
                this._diode.addClass('active');
            }
            return item;
        };

        Control.prototype._checkbox = function (screen) {
            let _this = this;
            let checked = this._key && screen.enabled(this._key);
            let checkbox = $('<input>', {class: 'checkbox', type: 'checkbox'})
                .prop('checked', checked)
                .change(function (ev, trg) {
                    let checked = this.checked;
                    _this._checkActive();
                    if (_this._key) {
                        screen.toggle(_this._key, checked);
                    }
                    if (_this._parent) {
                        _this._parent._checkParent(screen);
                        if (trg !== false) {
                            _this._checkLinked(screen, checked);
                        }
                    } else {
                        for (let child of _this._childs) {
                            child._check(screen, checked);
                            child._checkActive();
                            if (child._key) {
                                screen.toggle(child._key, checked);
                            }
                            if (trg !== false) {
                                child._checkLinked(screen, checked);
                            }
                        }
                    }
                });
            this._cbxes.set(screen, checkbox);
            this._checkboxes.append(checkbox);
            if (this._parent) {
                this._parent._checkParent(screen);
            }
            return checked;
        };

        Control.prototype._check = function (screen, checked) {
            this._cbxes.get(screen).prop('checked', checked);
        };

        Control.prototype._checkParent = function (screen) {
            let check = true;
            for (let child of this._childs) {
                check &= child._checked(screen);
            }
            this._check(screen, check);
            this._checkActive();
        };

        Control.prototype._checkActive = function () {
            let active = false;
            for (let checkbox of this._cbxes.values()) {
                active |= checkbox.prop('checked');
            }
            if (!active && this._diode.hasClass('active')) {
                this._diode.removeClass('active')
            } else if (active && !this._diode.hasClass('active')) {
                this._diode.addClass('active')
            }
        };

        Control.prototype._checked = function (screen) {
            return this._cbxes.get(screen) && this._cbxes.get(screen).prop('checked');
        };

        Control.prototype._checkLinked = function (screen, checked) {
            this._linked._check(screen, checked);
            this._linked._cbxes.get(screen).trigger('change', false);
        };

        return Control;
    })();

    let Controls = (function () {
        let _streams = [],
            _streamGroups = {},
            _streamsCount = 0,
            _nodes = [],
            _nodeGroups = {},
            _nodesCount = 0;

        return {
            handler: function (controls) {
                let link = {};

                $.each(controls.streams, function (stream, nodes) {
                    _streamsCount++;
                    let group = $('<div>', {class: 'list-group'});
                    let streamControl = new Control(STREAM, stream);
                    _streams.push(streamControl);
                    group.append(streamControl.render());
                    for (let node of nodes) {
                        let key = Key.of(node, stream);
                        let nodeControl = new Control(NODE, node, key, streamControl);
                        link[key] = nodeControl;
                        _streams.push(nodeControl);
                        group.append(nodeControl.render());
                    }
                    $('#streams').append(group);
                    _streamGroups[stream] = group;
                });

                $.each(controls.nodes, function (node, streams) {
                    _nodesCount++;
                    let group = $('<div>', {class: 'list-group'});
                    let nodeControl = new Control(NODE, node);
                    _nodes.push(nodeControl);
                    group.append(nodeControl.render());
                    for (let stream of streams) {
                        let key = Key.of(node, stream);
                        let streamControl = new Control(STREAM, stream, key, nodeControl);
                        streamControl.link(link[key]);
                        _nodes.push(streamControl);
                        group.append(streamControl.render());
                    }
                    $('#nodes').append(group);
                    _nodeGroups[node] = group;
                });
            },
            addScreen: function (screen) {
                for (let stream of _streams) {
                    stream.addScreen(screen);
                }
                for (let node of _nodes) {
                    node.addScreen(screen);
                }
            },
            delScreen: function (screen) {
                for (let stream of _streams) {
                    stream.delScreen(screen);
                }
                for (let node of _nodes) {
                    node.delScreen(screen);
                }
            },
            filterStreams: function (ev) {
                let value = this.value;
                $.each(_streamGroups, function (stream, group) {
                    group.toggle(stream.includes(value))
                })
            },
            filterNodes: function () {
                let value = this.value;
                $.each(_nodeGroups, function (node, group) {
                    group.toggle(node.includes(value))
                })
            },
            streamsCount: function () {
                return _streamsCount;
            },
            nodesCount: function () {
                return _nodesCount;
            }
        };
    })();

    let WebClient = (function () {
        let _socket,
            _handlers = {};

        return {
            connect: function () {
                _socket = new WebSocket('ws://' + window.location.host + '/ws');
                _socket.onmessage = function (ev) {
                    let data = JSON.parse(ev.data);
                    let handler = _handlers[data.type];
                    handler(data);
                }
            },
            handle: function (type, handler) {
                _handlers[type] = handler;
            }
        }
    })();

    let Stats = (function () {
        const STATS = $('#stats');
        let _started = new Date(),
            _stat = function (num, label) {
                STATS.append($('<div>', {class: 'stats'})
                    .append($('<span>', {class: 'stats-num', text: num}))
                    .append(' ')
                    .append($('<span>', {class: 'stats-label', text: label}))
                );
            };

        return {
            show: function () {
                STATS.empty();
                _stat(Controls.streamsCount(), 'Streams');
                _stat(Controls.nodesCount(), 'Nodes');
                _stat(Screens.messagesCount(), 'Messages');
                let elapsed = new Date() - _started;
                let min = Math.round(elapsed / 60000);
                let sec = Math.round(elapsed % 60000 / 1000);
                _stat(min + (sec < 10 ? ':0' : ':') + sec, 'elapsed');
            }
        }
    })();

    Screens.init();

    $('#streams-filter').on('input', Controls.filterStreams);
    $('#nodes-filter').on('input', Controls.filterNodes);

    $('#new-screen').click(function () {
        let screen = Screens.addScreen();
        if (screen) {
            Controls.addScreen(screen);
        }
    });

    WebClient.handle("controls", Controls.handler);
    WebClient.handle("message", Screens.handler);
    WebClient.connect();

    Stats.show();
    setInterval(Stats.show, 1000);
});