#!/usr/bin/env Python

def build_terms(query):
    terms = map(lambda x: x.strip(), filter(lambda x: x, query.strip().split(".")))
    ret = []
    for t in terms:
        ret.append(Term(map(lambda x: x.strip(), t.split(" "))))
    return ret
        

select = ["?vav"]
terms = [["?zone","rdf:type","brick:HVAC_Zone"],
         ["?vav", "bf:feeds","?zone"]]

class Term:
    def __init__(self, triple):
        self.subject = triple[0]
        self.predicate = triple[1]
        self.object = triple[2]
        self.subjectIsVar = self.subject.startswith("?")
        self.predicateIsVar = self.predicate.startswith("?")
        self.objectIsVar = self.object.startswith("?")
        self.subjectIsResolved = not self.subjectIsVar
        self.predicateIsResolved = not self.predicateIsVar
        self.objectIsResolved = not self.objectIsVar
        depends_on = []
        self.children = []
        self.parents = []
    def numberVars(self):
        return 1*(self.subjectIsVar and not self.subjectIsResolved) + \
               1*(self.predicateIsVar and not self.predicateIsResolved) +  \
               1*(self.objectIsVar and not self.objectIsResolved)
    def variables(self):
        v = []
        v.append(self.subject) if self.subjectIsVar else None
        v.append(self.predicate) if self.predicateIsVar else None
        v.append(self.object) if self.objectIsVar else None
        return v
    def getParentVariables(self):
        v = self.variables()
        for p in self.parents:
            print self, "get par",p
            v.extend(p.getParentVariables())
        return v
    def resolveAll(self):
        self.subjectIsResolved = True
        self.objectIsResolved = True
        self.predicateIsResolved = True
        
    def dependsOn(self, term):
        myvars = self.variables()
        depends = False
        if len(myvars) == 0:
            return depends
        for v in myvars:
            if v in term.variables():
                depends = True
                # TODO: loop detection?
                # problem is we get triples who are each others parents and it loops forever
                if term not in self.parents:
                    self.parents.append(term)
                break
        return depends

    def resolveWith(self, term):
        # pull in all resolved vars from terms parents too
        defined = term.getParentVariables()

        myvars = self.variables()
        if len(myvars) == 0:
            return
        for v in myvars:
            if v in defined:
                if v == self.subject:
                    self.subjectIsResolved = True
                elif v == self.predicate:
                    self.predicateIsResolved = True
                elif v == self.object:
                    self.objectIsResolved = True
        return

    def dump(self, indent=0):
        if indent > 5:
            return
        print "  "*indent + str(self)
        for child in self.children:
            child.dump(indent+1)
        
    def __repr__(self):
        return """<%s %s %s>""" % (self.subject, self.predicate, self.object)

def run_terms(terms):
    # first iterate through and find if there are any 'defining' terms: terms that only
    # have one variable
    defining_terms = []
    print terms
    for t in terms:
        if t.numberVars() == 1:
            t.resolveAll()        
            defining_terms.append(t)

    # now we find the defining terms that depend on these
    finished = False
    while not finished:
        finished = True
        for idx, t in enumerate(terms):
            if t.numberVars() == 0: continue
            if t.numberVars() < 2: 
                t.resolveAll()
                terms[idx] = t
                continue
            else:
                print t, t.numberVars()
            finished = False
            for idx2, dt in enumerate(terms):
                if t != dt and t.dependsOn(dt) :
                    print 'depends',t,dt,t.numberVars()
                    t.resolveWith(dt)
                    print 'resolved?',t,dt,t.numberVars()
                    terms[idx] = t
                    if t not in dt.children:
                        
                        dt.children.append(t)
                        terms[idx2] = dt

    print "DONE"
    defining_terms[0].dump()
    #for t in defining_terms:
    #    t.dump()
    return


#q1 = """
#    ?vav rdf:type brick:VAV .
#    ?zone rdf:type brick:HVAC_Zone .
#    ?room rdf:type brick:Room .
#    ?abc bf:feeds ?room .
#    ?abc rdf:type brick:AHU .
#"""
#run_terms(build_terms(q1))
#print '-'*30

#q1 = """
#    ?sensor rdf:type/rdfs:subClassOf* brick:CO2_Sensor .
#    ?room rdf:type brick:Room .
#    ?sensor bf:isPointOf ?room .
#"""
#run_terms(build_terms(q1))
#print '-'*30

#q1 = """
#    ?meter rdf:type brick:Power_Meter .
#    ?room rdf:type brick:Room .
#    ?sensor bf:isPointOf ?room .
#"""
#run_terms(build_terms(q1))
#
#print '-'*30
#
q1 = """
    ?meter rdf:type brick:Power_Meter .
    ?room rdf:type brick:Room .
    ?meter bf:isPointOf ?equipment .
    ?equipment rdf:type ?class .
    ?class rdfs:subClassOf+ brick:Heating_Ventilation_Air_Conditioning_System .
    ?zone rdf:type/rdfs:subClassOf* brick:HVAC_Zone .
    ?equipment bf:feeds+ ?zone .
    ?zone bf:hasPart ?room .
"""
run_terms(build_terms(q1))
#print '-'*30
#
#q1 = """
#    ?meter rdf:type/rdfs:subClassOf* brick:Power_Meter .
#    ?loc rdf:type ?loc_class .
#    ?loc_class rdfs:subClassOf+ brick:Location .
#
#    ?loc bf:hasPoint ?meter .
#"""
#run_terms(build_terms(q1))
#print '-'*30

#q1 = """
#    ?a bf:feeds ?b .
#    ?b bf:feeds ?c .
#    ?c bf:feeds ?d .
#    ?d bf:feeds ?d .
#    ?e bf:feeds ?loc .
#    ?loc bf:hasPoint brick:Power_Meter .
#"""
#run_terms(build_terms(q1))
#
