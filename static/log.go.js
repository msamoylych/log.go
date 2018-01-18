(function () {
    $.get("controls", parseControls);
    initWS();
})();

var STREAMS = $('#streams');
var NODES = $('#nodes');

var streamsCheckboxMap = new Map();
var nodesCheckboxMap = new Map();

function parseControls(data) {
    var streams = data.streams;
    var nodes = data.nodes;

    for (var stream in streams) {
        if (!streams.hasOwnProperty(stream)) {
            continue
        }

        var streamsGroup = div({id: stream, class: 'list-group'});
        var streamCheckboxMap = new Map();
        var streamNode = listGroupItem(stream, true);
        streamsGroup.append(streamNode);
        var nds = streams[stream];
        for (var n = 0; n < nds.length; n++) {
            streamsGroup.append(listGroupItem(nds[n], false, streamNode, streamCheckboxMap));
        }
        STREAMS.append(streamsGroup);
        streamsCheckboxMap.set(stream, streamCheckboxMap);
    }
    for (var node in nodes) {
        if (!nodes.hasOwnProperty(node)) {
            continue
        }

        var nodesGroup = div({id: node, class: 'list-group'});
        var nodeCheckboxMap = new Map();
        var nodeNode = listGroupItem(node, false);
        nodesGroup.append(nodeNode);
        var strms = nodes[node];
        for (var s = 0; s < strms.length; s++) {
            nodesGroup.append(listGroupItem(strms[s], true, nodeNode, nodeCheckboxMap));
        }
        NODES.append(nodesGroup);
        nodesCheckboxMap.set(node, nodeCheckboxMap);
    }
}

function listGroupItem(name, stream, parent, map) {
    var prnt = parent === undefined;
    var streams = (stream && prnt) || (!stream && !prnt);
    var item = div({class: ['list-group-item', prnt ? 'list-group-item-primary' : 'list-group-item-secondary']});
    item.append(div({class: ["diode", stream ? streamColor(name) : nodeColor(name)]}));
    item.append(div({text: name}));
    var checkboxes = div({class: "checkboxes"});
    var checkbox = $('<input type="checkbox" checked="checked">');
    checkbox.attr('name', name);
    var checkboxElement = checkbox[0];
    if (!prnt) {
        var m = streams ? streamsCheckboxMap : nodesCheckboxMap;
        if (!m.has(name)) {
            m.set(name, [])
        }
        m.get(name).push(checkboxElement);
        map.set(name, checkboxElement);
    }
    if (prnt) {
        jQuery.data(checkboxElement, 'children', []);
        checkbox.change(function (evt, trg) {
            var children = jQuery.data(this, 'children');
            for (var c = 0; c < children.length; c++) {
                $(children[c]).prop('checked', this.checked)
            }

            var map = streams ? nodesCheckboxMap : streamsCheckboxMap;
            var linkedCheckboxes = map.get(this.name);
            for (c = 0; c < linkedCheckboxes.length; c++) {
                $(linkedCheckboxes[c]).prop('checked', this.checked);
                if (trg !== false) {
                    $(linkedCheckboxes[c]).trigger('change', false)
                }
            }
        });
    } else {
        var parentCheckboxes = parent.children('.checkboxes')[0];
        var parentCheckboxElement = parentCheckboxes.childNodes[0];
        jQuery.data(checkboxElement, 'parent', parentCheckboxElement);
        jQuery.data(parentCheckboxElement, 'children').push(checkboxElement);
        checkbox.change(function (evt, trg) {
            var parent = jQuery.data(this, 'parent');
            if (this.checked) {
                var check = true;
                var children = jQuery.data(parentCheckboxElement, 'children');
                for (var c = 0; c < children.length; c++) {
                    if (!children[c].checked) {
                        check = false;
                        break
                    }
                }
                if (check) {
                    $(parent).prop('checked', true)
                }
            } else {
                $(parent).prop('checked', false)
            }

            var map = streams ? nodesCheckboxMap : streamsCheckboxMap;
            var linkedCheckbox = map.get(this.name).get(parentCheckboxElement.name);
            $(linkedCheckbox).prop('checked', this.checked);
            if (trg !== false) {
                $(linkedCheckbox).trigger('change', false)
            }
        })
    }
    checkboxes.append(checkbox);
    item.append(checkboxes);
    return item
}

var EVENTS = $('#events');

function initWS() {
    var ws = new WebSocket('ws://' + window.location.host + '/ws');
    ws.onmessage = function (ev) {
        var event = JSON.parse(ev.data);
        var row = div({class: 'event'});
        row.append(span({class: ['stream', streamColor(event.stream)], text: event.stream}));
        row.append(' ');
        row.append(span({class: ['node', nodeColor(event.node)], text: event.node}));
        row.append(' ');
        row.append(span({class: 'message', text: event.message}));
        EVENTS.append(row);
        EVENTS[0].scrollTop = EVENTS[0].scrollHeight;
    }
}

function filterStreams() {
    filter(STREAMS, $('#streams-filter').val())
}

function filterNodes() {
    filter(NODES, $('#nodes-filter').val())
}

function filter(list, pattern) {
    list.children().each(function () {
        if (pattern === '' || this.id.indexOf(pattern) !== -1) {
            $(this).show()
        } else {
            $(this).hide()
        }
    })
}

var streamsColors = {};

function streamColor(stream) {
    if (stream in streamsColors) {
        return streamsColors[stream]
    } else {
        var color = nextColor();
        streamsColors[stream] = color;
        return color;
    }
}

var nodesColors = {};

function nodeColor(node) {
    if (node in nodesColors) {
        return nodesColors[node]
    } else {
        var color = nextColor();
        nodesColors[node] = color;
        return color;
    }
}

var c = 0;

function nextColor() {
    c = c === 20 ? 1 : c + 1;
    return 'color' + c;
}

function div(props) {
    var div = $('<div></div>');
    if (props.id) {
        div.attr('id', props.id)
    }
    if (props.class) {
        if (Array.isArray(props.class)) {
            for (var i = 0; i < props.class.length; i++) {
                div.addClass(props.class[i])
            }
        } else {
            div.addClass(props.class)
        }
    }
    if (props.text) {
        div.text(props.text)
    }
    return div
}

function span(props) {
    var span = $('<span></span>');
    if (props.class) {
        if (Array.isArray(props.class)) {
            for (var i = 0; i < props.class.length; i++) {
                span.addClass(props.class[i])
            }
        } else {
            span.addClass(props.class)
        }
    }
    if (props.text) {
        span.text(props.text)
    }
    return span
}