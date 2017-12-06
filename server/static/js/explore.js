var textarea = document.getElementById("queryarea");
var cm = CodeMirror.fromTextArea(textarea, {
mode:  "application/sparql-query",
matchBrackets: true,
lineNumbers: true
});
cm.refresh();

var submit_query = function(query) {
  var html = "";
  var begin = moment();
  console.log(query);
  $("#errortext").hide();
  $.post("/api/queryclassdot", query, function(data) {

      var end = moment();
      var duration = moment.duration(end - begin);
      $("#elapsed").text(duration.milliseconds() + " ms");
      console.log(data);

      d3.select("#mynetwork svg").remove();

      var graph = DotParser.parse(data);
      console.log(graph);
      var width=1000;
      var height=500;
      var svg = d3.select("#mynetwork").append("svg")
      .attr("width", "100%")
      .attr("height", "100%")
      .call(d3.behavior.zoom().on("zoom", function () {
            svg.attr("transform", "translate(" + d3.event.translate + ")" + " scale(" + d3.event.scale + ")")
            }))
      .append("g");

      var edges = [];
      var nodes = [];
      var nodenames = [];
      var childs = graph.children;
      for (var i=0; i<childs.length; i++) {
        var e = childs[i];
        if (e.edge_list == null) {
          var n = e.node_id.id;
          if (nodenames.indexOf(n) == -1) {
            nodenames.push(n);
            nodes.push({name: n, color: e.attr_list[0].eq});
          } else {
            nodes[nodenames.indexOf(n)].color = e.attr_list[0].eq;
          }
          continue;
        }
        var n1 = e.edge_list[0].id;
        var n2 = e.edge_list[1].id;
        if (nodenames.indexOf(n1) == -1) {
          nodenames.push(n1);
          nodes.push({name: n1});
        }
        if (nodenames.indexOf(n2) == -1) {
          nodenames.push(n2);
          nodes.push({name: n2});
        }

        edges.push({
          source: nodenames.indexOf(n1),
          target: nodenames.indexOf(n2),
          name: e.attr_list[0].eq,
          left: false,
          right: true,
          weight: 1
        });
      }

      console.log(nodes);
      console.log(edges);

      var force = d3.layout.force()
        .gravity(.05)
        .distance(250)
        .charge(-1000)
        .size([width, height]);
      force
        .nodes(nodes)
        .links(edges)
        .start();

      var node = svg.selectAll(".node")
        .data(nodes)
        .enter().append("g")
        .attr("class", "node")
        .style("fill", function(d) {
            console.log(d);
            if (d.color != null) {
            return d.color;
            } else {
            return "#333";
            }
            })
      .call(force.drag);

      node.append("circle")
        .attr("r","10");

      node.append("text")
        .attr("dx", 12)
        .attr("dy", ".35em")
        .text(function(d) { return d.name; });

      // build the arrow.
      svg.append("svg:defs").selectAll("marker")
        .data(["end"])      // Different link/path types can be defined here
        .enter().append("svg:marker")    // This section adds in the arrows
        .attr("id", String)
        .attr("viewBox", "0 -5 10 10")
        .attr("refX", 15)
        .attr("refY", -1.5)
        .attr("markerWidth", 6)
        .attr("markerHeight", 6)
        .attr("orient", "auto")
        .append("svg:path")
        .attr("d", "M0,-5L10,0L0,5");

      // add the links and the arrows
      var path = svg.append("svg:g").selectAll("path")
        .data(edges)
        .enter().append("svg:path")
        .attr("class", "link")
        .attr("marker-end", "url(#end)")
        .attr("id", function(d,i) { return 'edge'+i;});

      var edgelabels = svg.selectAll(".edgelabel")
        .data(edges)
        .enter()
        .append('text')
        .style("pointer-events", "none")
        .attr({'class':'edgelabel',
            'id':function(d,i){return 'edgelabel'+i},
            'dx':50,
            'dy':-5,
            'font-size':18,
            'fill':'#aaa'});
      edgelabels.append('textPath')
        .attr('xlink:href',function(d,i) {return '#edge'+i})
        .style("pointer-events", "none")
        .text(function(d,i){return d.name})

      force.on("tick", function() {

          node.attr("transform", function(d) { return "translate(" + d.x + "," + d.y + ")"; });
          // handles to link and node element groups
          path.attr('d', function(d) {
              var deltaX = d.target.x - d.source.x,
              deltaY = d.target.y - d.source.y,
              dist = Math.sqrt(deltaX * deltaX + deltaY * deltaY),
              normX = deltaX / dist,
              normY = deltaY / dist,
              sourcePadding = 5;
              targetPadding = 5;
              sourceX = d.source.x + (sourcePadding * normX),
              sourceY = d.source.y + (sourcePadding * normY),
              targetX = d.target.x - (targetPadding * normX),
              targetY = d.target.y - (targetPadding * normY);
              return 'M' + sourceX + ',' + sourceY + 'L' + targetX + ',' + targetY;
              });

          edgelabels.attr('transform',function(d,i){
              if (d.target.x<d.source.x){
                bbox = this.getBBox();
                rx = bbox.x+bbox.width/2;
                ry = bbox.y+bbox.height/2;
                return 'rotate(180 '+rx+' '+ry+')';
               } else {
                return 'rotate(0)';
               }
          });
      });

  }).fail(function(e) {
    $("#errortext").show();
    $("#errortext > p").text(e.responseText);
    });
}

// init collapsible parts
$('.collapsible').collapsible();

$('.button-collapse').sideNav({
menuWidth: 350, // Default is 240
}
);

// https://davidwalsh.name/javascript-debounce-function
function debounce(func, wait, immediate) {
  var timeout;
  return function() {
    var context = this, args = arguments;
    var later = function() {
      timeout = null;
      if (!immediate) func.apply(context, args);
    };
    var callNow = immediate && !timeout;
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
    if (callNow) func.apply(context, args);
  };
};

var postponed_eval = debounce(function() {
    submit_query(cm.getValue())
    }, 500);

cm.on("change", postponed_eval);

// run once
var querytext = $("#queryarea").val();
submit_query(querytext);
