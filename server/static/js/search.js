var textarea = document.getElementById("searchtext");

$(document).ready(function() {
    var url = new URL(window.document.location);
    var klass = url.searchParams.get("class");
    var node = url.searchParams.get("node");
    if (klass != null) {
        list_instances(klass);
    }
    if (node != null) {
        render_node(node);
    }
});

var render_node = function(node) {
    var nodeQuery = "SELECT ?pred ?obj ?class WHERE {?obj rdf:type ?class . <" + node + "> ?pred ?obj };";
    var html = "<div class='row'><div class='col s12 m12'><div class='card darken-1'>";
    $.post("/api/query", nodeQuery, function(data) {
        console.log(data);
        html += "<div class='card-content'><span class='card-title'>" + node + "</span>";
        /*
         * TODO: We need to have the "summarize" in here because its possible to have an instance with a LOT of out-edges (e.g. an AHU connected to all of its VAVs).
         * Here, loop through all of the rows and group objects by predicate. Then we can render a sliding window that implements the "..." David was talking about.
         * Will need some materializecss magic here
         */
        var groupedEdges = {};
        data.Rows.forEach(function(row) {
            var pred = row['?pred'].Value;
            if (groupedEdges[pred] == null) {
                groupedEdges[pred] = [];
            }
            groupedEdges[pred].push(row);
        });
        var compound = [];
        var simple = [];
        for (var pred in groupedEdges) {
            if (groupedEdges[pred].length > 3) {
                compound.push(pred);
            } else {
                simple.push(pred);
            }
        }

        html += "<ul class='collection collapsible' data-collapsible='accordion'>";
        simple.forEach(function(pred) {
            var row = groupedEdges[pred][0];
            var pred = row['?pred'].Value;
            var full_object = row['?obj'].Namespace + '#' + row['?obj'].Value;
            var object = row['?obj'].Value;
            var klass = row['?class'].Value;
            html += "<li class='collection-item'>";
            html += "<div class='row'>";
                html += "<div class='col s4'><b>" + pred + "</b></div>";
                html += "<div class='col s4'>";
                    if (klass == "Class") {
                        html += "<a href='/search?class=" + encodeURIComponent("brick:"+object)+"'>" + object + "</a>";
                    } else {
                        html += "<a href='/search?node=" + encodeURIComponent(full_object)+"'>" + object + "</a>";
                    }
                html += "</div>";
                html += "<div class='col s4'>";
                    html += "<a href='/search?class="+ encodeURIComponent("brick:"+klass) + "'>" + klass + "</a>";
                html += "</div>";
            html += "</div>";
            html += "</li>";
        });

        compound.forEach(function(pred) {
            html += "<li>";
                html += "<div class='collapsible-header'><b>"+pred+"</b> (click to expand) </div>";
                html += "<div class='collapsible-body'>";
                html += "<ul class='collection'>";
                groupedEdges[pred].forEach(function(row) {
                    var pred = row['?pred'].Value;
                    var full_object = row['?obj'].Namespace + '#' + row['?obj'].Value;
                    var object = row['?obj'].Value;
                    var klass = row['?class'].Value;
                    html += "<li class='collection-item grey lighten-4'>";
                    html += "<div class='row'>";
                        html += "<div class='col s4'><b>" + pred + "</b></div>";
                        html += "<div class='col s4'>";
                            html += "<a href='/search?node=" + encodeURIComponent(full_object)+"'>" + object + "</a>";
                        html += "</div>";
                        html += "<div class='col s4'>";
                            html += "<a href='/search?class="+ encodeURIComponent("brick:"+klass) + "'>" + klass + "</a>";
                        html += "</div>";
                    html += "</div>";
                    html += "</li>";
                });
                html += "</ul>";
                html += "</div>";
            html += "</li>";
        });

        html += "</ul>";

        html += "</div></div></div></div>";
        $("#noderender").html(html);
        $('.collapsible').collapsible({
            accordian: true,
            onOpen: function(e) { console.log("open", e); },
        });
        
        console.log(data);
    });
    $('ul.tabs').tabs('select_tab', 'tab_traversal');
};

  var list_instances = function(klass) {
    if (klass.slice(6,10) == "http") {
        klass = 'brick:'+klass.slice(klass.indexOf('#')+1, klass.length);
    }
    var url = new URL(window.document.location);
    url.searchParams.set("class", klass)
    history.pushState(null,null,url);
    $('ul.tabs').tabs('select_tab', 'tab_instances');
    var instanceQuery = "SELECT ?inst ?class WHERE { ?inst rdf:type/rdfs:subClassOf* " + klass + " . ?inst rdf:type ?class };";
    var html = "<thead><tr><th>Instance</th><th>Class</th></tr></thead><tbody>";
    $.post("/api/query", instanceQuery, function(data) {
        console.log(data);
        if (data.Count > 0) {
            data.Rows.forEach(function(element) {
                var inst = element["?inst"].Value;
                var full_inst = element["?inst"].Namespace+'#'+element["?inst"].Value;
                var klass = element["?class"].Value;
                html += "<tr><td><a href='/search?node=" + encodeURIComponent(full_inst) + "'>"+ inst + "</a></td>";
                html += "<td><a href='/search?class=" + encodeURIComponent("brick:"+klass) + "'>" + klass + "</a></td></tr>";
            });
        }
        html += "</tbody>";
        $("#instancetable").html(html);
    });
  }

  var submit_query = function(query) {
    $("#errortext").hide();

    var html = "<tbody>";
    $.post("/api/search", JSON.stringify(query), function(data) {
        var url = new URL(window.document.location);
        url.searchParams.set("query", query.Query);
        // replace URL without reloading
        history.pushState(null,null,url);

        if (data == null) {
            $("#errortext").show();
            $("#errortext > p").text("No results");
        } else {
            data.forEach(function(element) {
                html += "<tr data-name='"+element+"'><td>";
                html += element;
                html += "</td></tr>";
            });
            html += "</tbody>";
            $("#resultstable").html(html);
            $('#resultstable tr').click(function() {
              console.log("CLICK", $(this).data('name'));
              list_instances($(this).data('name'));
            });
        }
    }).fail(function(e) {
        $("#errortext").show();
        $("#errortext > p").text(e.responseText);
    });
  }
  var url = new URL(window.document.location);
  var query = url.searchParams.get("query");
  if (query != null) {
      $("#searchtext").val(query);
      submit_query({'Query':query, 'Number': 500});
  }


  // init collapsible parts
  $('.collapsible').collapsible();

  $('.button-collapse').sideNav({
    menuWidth: 350, // Default is 240
    }
  );

  // handle querying
  //$("#searchtext").on("input", function(event) {
  //  var querytext = $(this).val();
  //  submit_query({'Query':querytext, 'Number': 500});
  //  event.preventDefault();
  //});
  $("form :input").change(function() {
    var querytext = $("#searchtext").val();
    submit_query({'Query':querytext, 'Number': 500});
    event.preventDefault();
  });
  $("#queryform").submit(function(event) {
    var querytext = $("#searchtext").val();
    submit_query({'Query':querytext, 'Number': 500});
    event.preventDefault();
  });

