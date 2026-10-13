"""Python APIs for OpenSchool Enterprise, FamineLens and Kindred.

Historical, educational and animal-welfare outputs are decision support. They retain sources,
uncertainty and human review rather than presenting automated conclusions as fact.
"""
from __future__ import annotations
import hashlib, math
from dataclasses import dataclass
from fastapi import FastAPI, HTTPException, Response
from pydantic import BaseModel, Field
app=FastAPI(title="Humanity Python API",version="3.0.0",docs_url="/docs",redoc_url=None)
class Driver(BaseModel):name:str;contribution:float=Field(ge=-1,le=1)
class Insight(BaseModel):score:int=Field(ge=0,le=100);confidence:float=Field(ge=0,le=1);summary:str;drivers:list[Driver];model:str="simulated-baseline-v3";requires_human_review:bool=True
@dataclass(frozen=True)
class ModelConfig:summary:str;drivers:tuple[str,...];base:int
MODELS={"openschool":ModelConfig("Weak prerequisites receive additional practice while curriculum coverage remains balanced.",( "mastery","curriculum balance","available time"),48),"faminelens":ModelConfig("Price stress and weak response capacity dominate the illustrative risk scenario; historical similarity is not causation.",( "food prices","rainfall","response capacity"),67),"kindred":ModelConfig("Urgent welfare reports should be reviewed first, while capacity and sensitive wildlife locations remain protected.",( "reported severity","time waiting","shelter capacity"),55),"terrazero":ModelConfig("Impact depends on hazard intensity, exposure and local response capacity; public alerts remain authoritative.",( "hazard intensity","population exposure","response capacity"),62),"nuclear-legacy":ModelConfig("Civil protection decisions require certified monitoring; this educational scenario communicates uncertainty only.",( "release index","distance","weather"),45),"ecosignal":ModelConfig("Network coverage is preserved while configurable power and idle airtime are reduced; wildlife safety requires calibrated field evidence.", ("coverage constraint","relative habitat proxy","energy activity"),42)}
def _fraction(seed,index):return int.from_bytes(hashlib.sha256(f"{seed}:{index}".encode()).digest()[:4],"big")/(2**32-1)
def insight_for(slug):
 cfg=MODELS.get(slug)
 if not cfg:raise KeyError(slug)
 raw=[.22+_fraction(slug,i)*.66 for i in range(len(cfg.drivers))];total=sum(raw)
 return Insight(score=cfg.base,confidence=round(.55+_fraction(slug,99)*.3,2),summary=cfg.summary,drivers=[Driver(name=n,contribution=round(v/total,3)) for n,v in zip(cfg.drivers,raw)])

SUBJECTS=["Mathematics","Science","Language & reading","People & society","Arts & making","Movement & wellbeing","Life skills"]
class SchoolPlanRequest(BaseModel):
 organization:str=Field(default="Community Learning Cooperative",max_length=100);curriculum:str=Field(default="Curriculum-neutral",max_length=80);role:str=Field(default="parent",pattern="^(parent|teacher|administrator)$");age:int=Field(ge=5,le=18);minutes_per_day:int=Field(ge=60,le=480);days:int=Field(ge=1,le=7);interests:list[str]=Field(default_factory=list,max_length=5);mastery:dict[str,float]=Field(default_factory=dict)
@app.post("/v1/openschool/plan")
def school_plan(request:SchoolPlanRequest):
 weak=sorted(SUBJECTS,key=lambda s:request.mastery.get(s,.5));blocks=max(2,request.minutes_per_day//35);plan=[]
 for day in range(request.days):
  sessions=[]
  for slot in range(blocks):
   subject=weak[(day+slot)%len(weak)] if slot<2 else SUBJECTS[(day*blocks+slot)%len(SUBJECTS)];interest=request.interests[(day+slot)%len(request.interests)] if request.interests else "the local community"
   sessions.append({"subject":subject,"minutes":request.minutes_per_day//blocks,"objective":f"Age {request.age}: build {subject.lower()} through {interest}","activity":f"Explore, practise and create one offline artifact about {interest}."})
  plan.append({"day":day+1,"sessions":sessions})
 coverage={s:sum(x["subject"]==s for d in plan for x in d["sessions"]) for s in SUBJECTS}
 return{"organization":request.organization,"curriculum":request.curriculum,"role":request.role,"plan":plan,"coverage":coverage,"offline_ready":True,"governance":{"content_status":"review required","audit_event":"plan.generated","child_data":"local-first; no advertising"},"safeguard":"Use open-licensed, age-reviewed resources under parent or educator supervision."}

FAMINES=[
 {"id":"india-1876","name":"Great Indian Famine","period":"1876–1878","region":"Southern and western India","summary":"Severe drought and crop failure interacted with high prices, market access, administrative policy and limited relief.","uncertainty":"Mortality estimates and causal weight differ across historical sources.","sources":[{"title":"Encyclopaedia Britannica — Famine","url":"https://www.britannica.com/science/famine"},{"title":"Our World in Data — Famines","url":"https://ourworldindata.org/famines"}]},
 {"id":"bengal-1943","name":"Bengal Famine","period":"1943","region":"Bengal","summary":"Wartime inflation, entitlement loss, distribution failures, crop damage and policy responses combined in a contested historical explanation.","uncertainty":"Historians disagree on the relative importance of aggregate availability and distribution/entitlement failures.","sources":[{"title":"Encyclopaedia Britannica — Bengal famine of 1943","url":"https://www.britannica.com/topic/Bengal-famine-of-1943"},{"title":"Our World in Data — Famines","url":"https://ourworldindata.org/famines"}]},
]
class FamineScenario(BaseModel):rainfall_stress:float=Field(ge=0,le=100);price_inflation:float=Field(ge=0,le=300);conflict_disruption:float=Field(ge=0,le=100);transport_access:float=Field(ge=0,le=100);response_capacity:float=Field(ge=0,le=100)
@app.get("/v1/faminelens/events")
def famine_events():return{"events":FAMINES,"method":"Sourced comparative timeline; no single-factor causal claim."}
@app.post("/v1/faminelens/analyze")
def famine_analyze(x:FamineScenario):
 raw={"rainfall stress":x.rainfall_stress*.24,"price inflation":min(100,x.price_inflation/2)*.29,"conflict disruption":x.conflict_disruption*.17,"transport constraint":(100-x.transport_access)*.14,"response gap":(100-x.response_capacity)*.16};score=min(100,round(sum(raw.values()),1));total=max(1,sum(raw.values()));drivers=[{"name":k,"contribution":round(v/total,3)} for k,v in sorted(raw.items(),key=lambda z:-z[1])];india=min(100,round(.42*x.rainfall_stress+.25*min(100,x.price_inflation/2)+.18*(100-x.response_capacity)+.15*(100-x.transport_access)));bengal=min(100,round(.14*x.rainfall_stress+.38*min(100,x.price_inflation/2)+.2*x.conflict_disruption+.16*(100-x.transport_access)+.12*(100-x.response_capacity)))
 return{"risk_index":score,"drivers":drivers,"historical_similarity":[{"event":"Great Indian Famine, 1876–1878","similarity":india},{"event":"Bengal Famine, 1943","similarity":bengal}],"uncertainty":{"lower":max(0,round(score-18)),"upper":min(100,round(score+18))},"notice":"Illustrative similarity, not a prediction or causal attribution. Field evidence and analyst approval are required."}

CASES=[{"id":"a1","kind":"companion","label":"Dog injured after flooding","severity":4,"waiting":35,"dependents":2,"x":18,"y":34},{"id":"a2","kind":"wildlife","label":"Displaced deer near road","severity":3,"waiting":80,"dependents":0,"x":62,"y":25},{"id":"a3","kind":"farm","label":"Cattle without water","severity":5,"waiting":20,"dependents":18,"x":75,"y":68},{"id":"a4","kind":"wildlife","label":"Sensitive nesting site disturbance","severity":2,"waiting":140,"dependents":4,"x":38,"y":73},{"id":"a5","kind":"companion","label":"Cats awaiting temporary foster","severity":2,"waiting":65,"dependents":6,"x":51,"y":48}]
SHELTERS=[{"id":"s1","name":"Community Animal Shelter","capacity":14,"x":28,"y":52},{"id":"s2","name":"Mobile Veterinary Team","capacity":6,"x":66,"y":46},{"id":"s3","name":"Wildlife Rehabilitation Centre","capacity":8,"x":48,"y":18}]
class KindredRequest(BaseModel):capacity_multiplier:float=Field(default=1,ge=.25,le=2);heat_risk:float=Field(default=30,ge=0,le=100);flood_risk:float=Field(default=20,ge=0,le=100)
@app.get("/v1/kindred/network")
def kindred_network():return{"cases":[{**c,"x":round(c["x"]/10)*10 if c["kind"]=="wildlife" else c["x"],"y":round(c["y"]/10)*10 if c["kind"]=="wildlife" else c["y"],"location_precision":"coarse" if c["kind"]=="wildlife" else "community"} for c in CASES],"shelters":SHELTERS}
@app.post("/v1/kindred/triage")
def kindred_triage(req:KindredRequest):
 ranked=[]
 for c in CASES:
  environmental=(req.heat_risk if c["kind"] in("farm","companion") else req.flood_risk)*.25;score=c["severity"]*100+c["waiting"]*.8+min(100,c["dependents"]*4)+environmental;ranked.append({**c,"score":round(score,1),"explanation":[f"reported severity {c['severity']}/5",f"waiting {c['waiting']} minutes",f"{c['dependents']} animals dependent",f"environmental pressure {round(environmental,1)}"],"requires_expert_review":True})
 ranked.sort(key=lambda c:(-c["score"],c["id"]));capacity=round(sum(s["capacity"] for s in SHELTERS)*req.capacity_multiplier);return{"ranked":ranked,"available_capacity":capacity,"unallocated_animals":max(0,sum(max(1,c["dependents"]) for c in CASES)-capacity),"policy":"Priority supports human review; it is not a veterinary diagnosis or automated enforcement decision."}

@app.middleware("http")
async def security_headers(request,call_next):response:Response=await call_next(request);response.headers["X-Content-Type-Options"]="nosniff";response.headers["Referrer-Policy"]="no-referrer";return response
@app.get("/healthz")
def healthz():return{"status":"ok","projects":6}
@app.get("/v1/insights/{slug}",response_model=Insight)
def get_insight(slug:str):
 try:return insight_for(slug)
 except KeyError:raise HTTPException(404,"unknown project") from None
