from services.intelligence.app.main import *
def test_six_models():
 assert len(MODELS)==6
 for slug in MODELS:
  x=insight_for(slug);assert 0<=x.score<=100 and x.requires_human_review;assert abs(sum(d.contribution for d in x.drivers)-1)<.01

def test_openschool_enterprise_plan():
 r=school_plan(SchoolPlanRequest(organization="Learning Coop",curriculum="Open",role="teacher",age=12,minutes_per_day=180,days=7,interests=["water"],mastery={"Mathematics":.2}));assert len(r["plan"])==7;assert r["organization"]=="Learning Coop";assert all(r["coverage"][s]>0 for s in r["coverage"]);assert r["governance"]["content_status"]=="review required"

def test_famine_lens_sourced_and_monotonic():
 assert {e["period"] for e in FAMINES}=={"1876–1878","1943"};assert all(e["sources"] for e in FAMINES)
 low=famine_analyze(FamineScenario(rainfall_stress=10,price_inflation=10,conflict_disruption=5,transport_access=90,response_capacity=90));high=famine_analyze(FamineScenario(rainfall_stress=90,price_inflation=200,conflict_disruption=80,transport_access=20,response_capacity=10));assert high["risk_index"]>low["risk_index"];assert high["uncertainty"]["lower"]<high["uncertainty"]["upper"]

def test_kindred_hides_wildlife_locations_and_requires_review():
 network=kindred_network();assert all(c["location_precision"]=="coarse" for c in network["cases"] if c["kind"]=="wildlife")
 r=kindred_triage(KindredRequest(capacity_multiplier=.5,heat_risk=80,flood_risk=70));assert all(c["requires_expert_review"] for c in r["ranked"]);assert r["ranked"][0]["severity"]>=4;assert r["unallocated_animals"]>=0
