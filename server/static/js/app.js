var bus = new Vue({
    data: {
        loaded: false,
        headers: [],
        items: [],
    }
})

Vue.component('graph', {
    created: function() {
        submit_query();
    },
    template: '\
        <div id="mynetwork"></div>\
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
          }
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
        var q = to_query().split('.').join(".\n")
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
                for (key in r) {
                    r[key] = r[key].Value;
                }
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
    template: '\
        <span>\
            <textarea id="renderquery"></textarea>\
            <div id="renderresults">\
                <v-data-table v-if="bus.loaded" :pagination.sync="pagination" :headers="bus.headers" :items="bus.items" class="elevation-1">\
                    <template slot="items" slot-scope="props">\
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
