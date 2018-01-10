//var textarea = document.getElementById("queryarea");
//var cm = CodeMirror.fromTextArea(textarea, {
//  mode:  "application/sparql-query",
//  matchBrackets: true,
//  lineNumbers: true,
//  size: 100
//});
//cm.refresh();

https://stackoverflow.com/questions/1349404/generate-random-string-characters-in-javascript
// dec2hex :: Integer -> String
function dec2hex (dec) {
  return ('0' + dec.toString(16)).substr(-2)
}

// generateVar :: Integer -> String
function generateVar (len) {
  var arr = new Uint8Array((len || 40) / 2)
  window.crypto.getRandomValues(arr)
  return "?"+Array.from(arr, dec2hex).join('')
}

var QUERY = {
    "brick:Zone_Temperature_Sensor": {
        SELECT: "?room",
        WHERE: ["?room rdf:type brick:Room . "],
    },
};

var to_query = function() {
    var build = "SELECT";
    for (key in QUERY) {
        build += " " + QUERY[key].SELECT;
    }
    build += " WHERE { "
    for (key in QUERY) {
        build += " " + QUERY[key].WHERE.join(' ');
    }
    build += " };";
    return build;
}

var get_vars = function() {
    var build = [];
    for (key in QUERY) {
        build.push(QUERY[key].SELECT);
    }
    return build;
}


var find_edge_by_id = function(n, edgeid) {
    for (var i in n.edges) {
        if (n.edges[i].id == edgeid) {
            return n.edges[i];
        }
    }
    console.log('could not find', edgeid,'in',n);
    return null;
}

var find_node_by_id = function(n, nodeid) {
    for (var i in n.nodes) {
        if (n.nodes[i].id == nodeid) {
            return n.nodes[i];
        }
    }
    console.log('could not find', edgeid,'in',n);
    return null;
}


var get_var_name = function(name) {
    var split = name.split('|');
    if (split.length == 1) {
        return name;
    }
    return split[0];
}
var get_old_name = function(name) {
    var split = name.split('|');
    if (split.length == 1) {
        return '';
    }
    return split[1];
}

var submit_query = function() {
  var html = "";
  var begin = moment();
  $("#errortext").hide();
  var parsedData = {nodes: [], edges: []};
  $.post("/api/queryclassdot", to_query(), function(data) {
      if (network != null) {
          network.destroy();
      }
      var end = moment();
      var duration = moment.duration(end - begin);
      $("#elapsed").text(duration.milliseconds() + " ms");
      var newdata = vis.network.convertDot(data)
      parsedData.options = newdata.options;
      for (var idx in newdata.nodes) {
        var n = newdata.nodes[idx];
        if (get_var_name(n.id).length < n.id.length) {
            n.varname = get_var_name(n.id);
            n.id = get_old_name(n.id);
            if (n.id == 'bf:uri') {
                continue;
            }
            if (n.id == 'bf:uuid') {
                continue;
            }
            var found = false;
            parsedData.nodes.forEach(function(nn, idxx) {
                if (nn.id == n.id) {
                    parsedData.nodes[idxx].varname = n.varname;
                    parsedData.nodes[idxx].label = n.label;
                    parsedData.nodes[idxx].color = n.color;
                    found = true;
                }
            });
            if (!found) {
                parsedData.nodes.push(n);
            }
        } else {
            var dup = parsedData.nodes.find(function(dup) {
                return dup.id == n.id;
            });
            if (dup == null) {
                parsedData.nodes.push(n);
            }
        }
      }
      //parsedData.nodes = newnodes;
      for (var idx in newdata.edges) {
        var e = newdata.edges[idx];
        if (get_var_name(e.from).length < e.from.length) {
            e.from = get_old_name(e.from);
        }
        if (get_var_name(e.to).length < e.to.length) {
            e.to = get_old_name(e.to);
        }
        if (e.to == 'bf:uri') {
            e.to = generateVar(10);
            e.label = 'bf:uri';
            var n = {id: e.to, label: 'URI'};
            parsedData.nodes.push(n);
        } else if (e.to == 'bf:uuid') {
            e.to = generateVar(10);
            e.label = 'bf:uuid';
            var n = {id: e.to, label: 'UUID'};
            parsedData.nodes.push(n);
        }
        parsedData.edges.push(e);
      }

      var container = document.getElementById('mynetwork');
      var data = {
        nodes: parsedData.nodes,
        edges: parsedData.edges
      };
      var options = parsedData.options;
      options.interaction = {
        hover: true,
        selectable: true
      };
      options.layout = {
          hierarchical: {
            enabled: true,
            blockShifting: true,
            levelSeparation: 200,
            nodeSpacing: 100,
            edgeMinimization: true,
            direction: 'LR'
          }
      };
      //options.physics = {
      //  barnesHut: {
      //      //gravitationalConstant: -3000,
      //      springLength: 300,
      //      //avoidOverlap: .3,
      //  },
      //  timestep: 1
      //};

      var network = new vis.Network(container, data, options);
      network.on("click", function(params) {
        var clicked = network.getSelectedNodes()[0]
        if (clicked in QUERY) {
            delete QUERY[clicked];
            submit_query();
            return;
        }
        var edge = find_edge_by_id(parsedData, network.getSelectedEdges()[0]);
        var newclass = edge.to;
        var newvar = generateVar(5);
        var orignode = find_node_by_id(parsedData, edge.from);
        var clickednode = find_node_by_id(parsedData, edge.to);

        console.log("clicked", clicked);
        if (clickednode.label == "URI") {
            QUERY[newclass] = {
                SELECT: newvar,
                WHERE: [orignode.varname + " bf:uri " + newvar + " . "]
            };
            return;
        } else if (clickednode.label == "UUID") {
            QUERY[newclass] = {
                SELECT: newvar,
                WHERE: [orignode.varname + " bf:uuid " + newvar + " . "]
            };
            return;
        } else {
            var line1 = orignode.varname + " " + edge.label + " " + newvar + " . ";
            var line2 = newvar + " rdf:type " + newclass + " . ";
            var line3 = newvar + " " + generateVar(5) + " " + generateVar(5) + " . ";
            QUERY[newclass] = {
                SELECT: newvar,
                WHERE: [line1, line2, line3],
            }
        }
        console.log("NEW",to_query());
        submit_query();

      });
      network.redraw();
  }).fail(function(e) {
      $("#errortext").show();
      $("#errortext > p").text(e.responseText);
  });
}


//cm.on("change", function(e, x) {
//  //submit_query(cm.getValue());
//});

// run once
var querytext = $("#queryarea").val();

//submit_query();

