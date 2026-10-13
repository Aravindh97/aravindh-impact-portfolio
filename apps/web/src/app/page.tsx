import { ProductGrid } from "@/components/product-grid";
import { products } from "@/data/products";

export default function Home() {
  return <main id="main">
    <section className="hero">
      <div className="eyebrow">Python + Go + AI · systems for public good</div>
      <h1>Technology should expand what humanity can <em>protect, learn and imagine.</em></h1>
      <div className="hero-copy">
        <p>I'm <strong>Aravindh Varadharaj</strong>, moving from workflow automation at Amazon into full-stack software engineering. This portfolio spans enterprise open learning, historical famine analysis, animal welfare, multi-hazard disaster impact, nuclear civil safety and wildlife-aware network efficiency.</p>
        <div className="hero-actions"><a className="primary" href="#systems">View my projects</a><a href="/resume.pdf">Download my resume</a></div>
      </div>
      <div className="proof"><div><b>{products.length}</b><span>humanity systems</span></div><div><b>3</b><span>production languages</span></div><div><b>2</b><span>backend services</span></div><div><b>1</b><span>shared mission</span></div></div>
    </section>

    <section id="systems" className="section">
      <div className="section-heading"><span>01 · Portfolio</span><div><h2>One system for each frontier.</h2><p>Not another collection of clones. Each product begins with a real decision, identifies who could be harmed, and makes the path from prototype to production explicit.</p></div></div>
      <ProductGrid products={products} />
    </section>

    <section id="stack" className="section architecture">
      <div className="section-heading"><span>02 · Technical skills</span><div><h2>Modern frontend. Reliable services behind it.</h2><p>The new platform removes browser WebAssembly. Next.js handles the experience; Go owns high-concurrency operational APIs; Python owns modelling and intelligence.</p></div></div>
      <div className="architecture-grid">
        <article><small>Experience</small><h3>Next.js 16 + React 19</h3><p>TypeScript 7, App Router, server rendering, responsive accessible components and Tailwind CSS.</p></article>
        <article><small>Operational core</small><h3>Go services</h3><p>Public hazard ingestion, geospatial impact zones, civil-safety scenarios and coverage-constrained wireless-efficiency optimization over versioned HTTP APIs.</p></article>
        <article><small>Intelligence</small><h3>Python + FastAPI</h3><p>Governed learning plans, sourced historical comparison, welfare triage, uncertainty communication and human-review boundaries.</p></article>
        <article><small>Production path</small><h3>Postgres + observability</h3><p>PostGIS/pgvector where justified, Redis for coordination, OpenTelemetry, containers and private data services.</p></article>
      </div>
    </section>

    <section id="profile" className="section profile-section">
      <div className="section-heading"><span>03 · Experience</span><div><h2>From operational automation to full-stack engineering.</h2><p>I bring Amazon domain experience, a portfolio of tested systems, and a deliberate path toward production ownership.</p></div></div>
      <div className="profile-grid">
        <article><small>Experience</small><h3>Sr. Digital Marketing Associate · Amazon</h3><b>2021—Present</b><ul><li>Developed and automated internal tools and workflows, replacing repetitive manual work with reusable solutions.</li><li>Worked with stakeholders to identify process gaps, then designed, built and supported team automation.</li><li>Built full-stack Python and Go systems outside my core role to move into software engineering.</li></ul><small>Education</small><h3>B.E. Mechanical Engineering</h3></article>
        <article><small>Core engineering</small><h3>Python · Go · TypeScript</h3><p>REST APIs, concurrent systems, FastAPI, Next.js, React, data modelling, event sourcing, CRDTs, idempotency and testing.</p><small>Data &amp; intelligence</small><h3>Applied ML</h3><p>NumPy, pandas, logistic regression, panel regression, Bayesian Knowledge Tracing, retrieval and explainability.</p></article>
        <article><small>What I’m seeking</small><h3>Purpose-driven collaboration</h3><p>I’m eager to join a purpose-driven team, collaborate with its community, and build technology that creates meaningful, measurable improvements in people’s everyday lives.</p><div className="profile-actions"><a href="/resume.pdf">Download resume PDF</a><a href="https://github.com/Aravindh97" target="_blank" rel="noreferrer">GitHub</a><a href="https://www.linkedin.com/in/aravindh-varadharaj-253231201" target="_blank" rel="noreferrer">LinkedIn</a><a href="mailto:aravindhvaradha@gmail.com">Email me</a></div></article>
      </div>
    </section>

    <section className="section invitation">
      <span>04 · Purpose</span><h2>I&apos;m eager to build useful technology with communities that care about everyday impact.</h2>
      <div><a href="mailto:aravindhvaradha@gmail.com">aravindhvaradha@gmail.com</a><a href="#systems">Back to my projects ↑</a></div>
    </section>
  </main>;
}
