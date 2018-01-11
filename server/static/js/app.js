var bus = new Vue({
    data: {
        loaded: false,
        headers: [],
        items: [],
        searchterm: "",
        selected: [],
    }
})

Vue.component('graph', {
    created: function() {
        submit_query();
    },
    methods: {
        research: function(e) {
            console.log("CHANGE", e);
            QUERY = {}
            get_classes(e, function(res) {
                res.forEach(function(klass) {
                    var vn = generateVar(5);
                    QUERY[klass] = {SELECT: vn, WHERE: [vn + " rdf:type " + klass + " . "]};
                });
                console.log(to_query());
                //rebuildquery(e);
                submit_query();
            });
        },
    },
    template: '\
		<span>\
			<v-text-field name="brickclass" label="Brick class search" :value="bus.searchterm" class="input-group--focused" @input="research" single-line></v-text-field>\
			<div id="mynetwork"></div>\
		</span>\
    '
})

/*
 *        headers: [
 *                  {
 *                              text: 'Dessert (100g serving)',
 *                                          align: 'left',
 *                                                      sortable: false,
 *                                                                  value: 'name'
 *                                                                            },
 */

Vue.component('query', {
    data: function() {
      return {
          pagination: {
                rowsPerPage: 1000
          },
          selected: [],
      }
    },
    computed:  {
        selectvars: function() {
            return get_vars();
        },
    },
    mounted: function() {
        console.log("query", QUERY);
        var textarea = document.getElementById("renderquery");
        console.log(textarea);
        var cm = CodeMirror.fromTextArea(textarea, {
            mode:  "application/sparql-query",
            matchBrackets: true,
            lineNumbers: true,
        });
        var q = to_query_no_explore().split('.').join(".\n")
        q = q.split('{').join("{\n");
        q = q.split('}').join("}\n");
        cm.setValue(q);
        cm.refresh();
        bus.loaded = false;
        bus.headers = [];
        bus.items = [];
        axios.post('/api/query', to_query())
        .then(function(response) {
            console.log('response', response);
            // response.data.Rows
            bus.headers.push({
                text: 'plot',
                align: 'left',
                sortable: false,
            })
            get_vars().forEach(function(vn) {
                bus.headers.push({
                    text: vn,
                    align: 'right',
                    sortable: true,
                    value: vn,
                });
            });
            console.log(bus.headers);

            response.data.Rows.forEach(function(r) {
                var rowid = [];
                for (key in r) {
                    r[key] = r[key].Value;
                    if (key.indexOf('uuid') > 0) {
                        rowid.push(r[key]);
                    }
                }
                r.rowid = rowid;
                console.log("rowid", r.rowid);
                r.selected = false;
                bus.items.push(r);
            });
            bus.loaded = true;
        })
        .catch(function(error) {
            console.log('error', error);
        })
    },
    computed: {
        querytext: function() {
            return to_query();
        }
    },
    methods: {
        ping: function(e) {
            console.log("slected", bus.items.filter(row => row.selected));
        },
        plot: function(doall) {
            var streams = [];
            bus.items.map(function(row) {
                if (doall || row.selected) {
                    for (i in row.rowid) {
                        streams.push(row.rowid[i]);        
                    }
                }
            });
            axios.post('/permalink', JSON.stringify({'URL': 'https://plot.xbos.io', 'UUIDs': streams}))
            .then(function(response) {
                console.log(response);
                console.log("PERMALINK", 'https://plot.xbos.io/?'+response.data);
                window.open('https://plot.xbos.io/?'+response.data);
            })
            .catch(function(error) {
                console.log('error', error);
            })
        },
    },
    template: '\
        <span>\
            <textarea id="renderquery"></textarea>\
            <v-btn @click="plot(false)" color="green lighten-1" dark>Plot Selected</v-btn>\
            <v-btn @click="plot(true)" color="green lighten-1" dark>Plot All</v-btn>\
            <div id="renderresults" @click="ping">\
                <v-data-table v-if="bus.loaded" :pagination.sync="pagination" :headers="bus.headers" :items="bus.items" class="elevation-1">\
                    <template slot="items" slot-scope="props">\
                        <td><v-checkbox primary hide-details v-model="props.item.selected"></v-checkbox></td>\
                        <td v-for="selectVar in get_vars()" :key="selectVar" class="text-xs-right">{{ props.item[selectVar] }}</td>\
                    </template>\
                </v-data-table>\
                <p v-else>no data</p>\
            </div>\
        </span>\
    '
})

var vm = new Vue({
    el: '#app',
    data: {
        page: 'graph',
    },
    created: function() {
        console.log("hey");
        submit_query();
        //console.log(QUERY);
    },
    methods: {
        dograph: function() {
            console.log("GRAPH");
            this.page = 'graph';
        },
        doquery: function() {
            console.log("QUERY");
            this.page = 'query';
        }
    },
    computed: {
        render_graph: function() {
            return this.page == 'graph';
        },
        render_query: function() {
            return this.page == 'query';
        }
    },
    template: '\
        <v-app>\
            <v-content class="container">\
                <h1>Query Builder</h1>\
                <div class="text-xs-left">\
                    <v-btn @click="dograph" color="green lighten-1" dark>1. Graph</v-btn>\
                    <v-btn @click="doquery" color="blue lighten-1" dark>2. Query</v-btn>\
                </div>\
                <graph v-if="render_graph"></graph>\
                <query v-if="render_query"></query>\
            </v-content>\
        </v-app>\
    ',
})
