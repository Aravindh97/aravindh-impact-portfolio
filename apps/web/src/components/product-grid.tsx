"use client";

import { useMemo, useState } from "react";
import type { Product } from "@/data/products";

export function ProductGrid({ products }: { products: Product[] }) {
  const [filter, setFilter] = useState<"All" | Product["backend"]>("All");
  const visible = useMemo(() => filter === "All" ? products : products.filter((p) => p.backend.includes(filter)), [filter, products]);
  return <>
    <div className="filter-bar" role="group" aria-label="Filter systems by backend">
      {(["All", "Python", "Go"] as const).map((item) => <button key={item} type="button" aria-pressed={filter === item} onClick={() => setFilter(item)}>{item}</button>)}
      <span>{visible.length} systems</span>
    </div>
    <div className="product-grid">
      {visible.map((product) => <article className="product-card" key={product.slug} style={{ "--accent": product.accent } as React.CSSProperties}>
        <div className="product-top"><span>{product.index}</span><span>{product.category}</span></div>
        <h3>{product.name}</h3>
        <p className="promise">{product.promise}</p>
        <p>{product.outcome}</p>
        <div className="badges"><span>{product.backend}</span><span>{product.maturity}</span></div>
        <a href={`/systems/${product.slug}`}>Explore the project <span aria-hidden="true">→</span></a>
      </article>)}
    </div>
  </>;
}
