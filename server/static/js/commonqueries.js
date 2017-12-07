
$("#query_vavselect").click(function(event) {
    cm.setValue("SELECT ?vav WHERE {\n\t?vav rdf:type brick:VAV \n};");
    cm.setSize("100%", Math.max(300, cm.lineCount()*25));
    event.preventDefault();
    });
$("#query_tempselect").click(function(event) {
    cm.setValue("SELECT ?sensor ?room\nWHERE {\n\t?sensor rdf:type/rdfs:subClassOf* brick:Zone_Temperature_Sensor .\n\t?room rdf:type brick:Room .\n\t?vav rdf:type brick:VAV .\n\t?zone rdf:type brick:HVAC_Zone .\n\n\t?vav bf:feeds+ ?zone .\n\t?zone bf:hasPart ?room \n\t\n\t{ ?sensor bf:isPointOf ?vav }\n\tUNION\n\t{ ?sensor bf:isPointOf ?room }\n\n};");
    cm.setSize("100%", Math.max(300, cm.lineCount()*25));
    cm.refresh();
    event.preventDefault();
    });
$("#query_vavcmd").click(function(event) {
    cm.setValue("SELECT ?vlv_cmd ?vav\nWHERE {\n\t{ ?vlv_cmd rdf:type brick:Reheat_Valve_Command }\n\tUNION\n\t{ ?vlv_cmd rdf:type brick:Cooling_Valve_Command }\n\t?vav rdf:type brick:VAV .\n\t?vav bf:hasPoint+ ?vlv_cmd \n};");
    cm.setSize("100%", Math.max(300, cm.lineCount()*25));
    event.preventDefault();
    });
$("#query_floors").click(function(event) {
    cm.setValue("SELECT ?floor ?room ?zone\nWHERE {\n\t?floor rdf:type brick:Floor .\n\t?room rdf:type brick:Room .\n\t?zone rdf:type brick:HVAC_Zone .\n\t?room bf:isPartOf+ ?floor .\n\t?room bf:isPartOf+ ?zone \n};");
    cm.setSize("100%", Math.max(300, cm.lineCount()*25));
    event.preventDefault();
    });

