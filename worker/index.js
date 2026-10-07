/**
 * Cloudflare Worker: scripts.hhk.my.id
 * High-performance Reverse Proxy for https://github.com/hhkmy/scripts
 * 
 * Features:
 * - Direct plain-text bash script serving (no redirects required)
 * - Intelligent User-Agent routing:
 *   - Terminal (curl/wget): Displays clean CLI banner & script catalog
 *   - Web Browsers: Redirects to GitHub repository
 * - Cloudflare Edge Caching (5 minutes)
 * - Safe error handling & structured observability logging
 * - Support for GET, HEAD, and CORS preflight OPTIONS
 */

const GITHUB_REPO = "hhkmy/scripts";
const GITHUB_BRANCH = "main";
const UPSTREAM_BASE = `https://raw.githubusercontent.com/${GITHUB_REPO}/${GITHUB_BRANCH}`;

// Known scripts catalog for terminal welcome banner
const SCRIPTS_CATALOG = [
  { name: "antigravity", desc: "Google Antigravity 2.0 Updater & Installer (Linux)" },
  { name: "go",          desc: "Go Compiler & Version Manager (Linux & macOS)" },
  { name: "gohugo",      desc: "Hugo Extended Static Site Generator (with Sass/SCSS)" }
];

function generateCliBanner() {
  const pad = (str, len) => str + " ".repeat(Math.max(0, len - str.length));
  const scriptList = SCRIPTS_CATALOG
    .map(s => `  • ${pad(s.name, 14)} - ${s.desc}`)
    .join("\n");

  return `
╭──────────────────────────────────────────────────────────╮
│   🛠️  hhkmy Scripts Hub — https://scripts.hhk.my.id      │
╰──────────────────────────────────────────────────────────╯

Usage:
  curl -fsSL https://scripts.hhk.my.id/<script> | bash

Available Scripts:
${scriptList}

Examples:
  curl -fsSL https://scripts.hhk.my.id/antigravity | bash
  curl -fsSL https://scripts.hhk.my.id/antigravity | sudo bash -s -- --global
  curl -fsSL https://scripts.hhk.my.id/go | bash
  curl -fsSL https://scripts.hhk.my.id/gohugo | bash -s -- --extended

GitHub Repository:
  https://github.com/${GITHUB_REPO}
\n`;
}

export default {
  async fetch(request, env, ctx) {
    try {
      const url = new URL(request.url);
      const path = url.pathname.replace(/^\/+/, ""); // strip leading slash
      const userAgent = (request.headers.get("user-agent") || "").toLowerCase();
      const isCli = userAgent.includes("curl") || userAgent.includes("wget") || userAgent.includes("httpie");

      // 1. CORS Preflight (OPTIONS)
      if (request.method === "OPTIONS") {
        return new Response(null, {
          status: 204,
          headers: {
            "Access-Control-Allow-Origin": "*",
            "Access-Control-Allow-Methods": "GET, HEAD, OPTIONS",
            "Access-Control-Max-Age": "86400"
          }
        });
      }

      // 2. Filter allowed HTTP methods
      if (request.method !== "GET" && request.method !== "HEAD") {
        console.warn(`[Method Not Allowed] ${request.method} on ${url.pathname}`);
        return new Response("405 Method Not Allowed\n", {
          status: 405,
          headers: {
            "Allow": "GET, HEAD, OPTIONS",
            "Content-Type": "text/plain; charset=utf-8"
          }
        });
      }

      // 3. Handle Root (/)
      if (!path) {
        if (isCli) {
          console.log("[Route] CLI Banner served");
          const banner = generateCliBanner();
          return new Response(request.method === "HEAD" ? null : banner, {
            status: 200,
            headers: {
              "Content-Type": "text/plain; charset=utf-8",
              "Cache-Control": "public, max-age=3600"
            }
          });
        }
        console.log(`[Route] Web Redirect to https://github.com/${GITHUB_REPO}`);
        return Response.redirect(`https://github.com/${GITHUB_REPO}`, 302);
      }

      // 4. Ignore static meta requests
      if (path === "favicon.ico" || path === "robots.txt") {
        return new Response(null, { status: 404 });
      }

      // 5. Security: Prevent traversal or hidden files
      if (path.includes("..") || path.startsWith(".") || path.includes("/")) {
        console.warn(`[Security Alert] Invalid path requested: ${path}`);
        return new Response("Invalid script path\n", {
          status: 400,
          headers: { "Content-Type": "text/plain; charset=utf-8" }
        });
      }

      // 6. Check Cloudflare Edge Cache
      const cacheKey = new Request(url.toString(), request);
      const cache = caches.default;
      let cachedResponse = await cache.match(cacheKey);

      if (cachedResponse) {
        console.log(`[Cache HIT] ${path}`);
        if (request.method === "HEAD") {
          return new Response(null, {
            status: cachedResponse.status,
            headers: cachedResponse.headers
          });
        }
        return cachedResponse;
      }

      console.log(`[Cache MISS] Fetching upstream: ${path}`);

      // 7. Fetch from GitHub upstream
      const upstreamUrl = `${UPSTREAM_BASE}/${path}`;
      const upstreamResponse = await fetch(upstreamUrl, {
        headers: {
          "User-Agent": "scripts.hhk.my.id Cloudflare-Worker"
        }
      });

      if (upstreamResponse.status === 404) {
        console.warn(`[404 Not Found] Script '${path}' does not exist upstream.`);
        return new Response(
          `Error: Script '${path}' not found in ${GITHUB_REPO}.\n\n` +
          `Visit https://scripts.hhk.my.id or https://github.com/${GITHUB_REPO} for available scripts.\n`,
          {
            status: 404,
            headers: { "Content-Type": "text/plain; charset=utf-8" }
          }
        );
      }

      if (!upstreamResponse.ok) {
        console.error(`[Upstream Error] ${upstreamResponse.status} ${upstreamResponse.statusText} on ${upstreamUrl}`);
        return new Response(`Upstream GitHub error: ${upstreamResponse.status} ${upstreamResponse.statusText}\n`, {
          status: 502,
          headers: { "Content-Type": "text/plain; charset=utf-8" }
        });
      }

      // 8. Construct response
      const body = await upstreamResponse.text();
      const headers = {
        "Content-Type": "text/plain; charset=utf-8",
        "Cache-Control": "public, max-age=300, s-maxage=300", // 5 minutes edge cache
        "X-Content-Type-Options": "nosniff",
        "Access-Control-Allow-Origin": "*",
        "X-Robots-Tag": "noindex"
      };

      const response = new Response(request.method === "HEAD" ? null : body, {
        status: 200,
        headers
      });

      // Cache full GET responses asynchronously
      if (request.method === "GET") {
        ctx.waitUntil(cache.put(cacheKey, response.clone()));
      }

      return response;
    } catch (err) {
      // 9. Unhandled Error: Log to Cloudflare Observability & return clean error
      console.error("[Worker Unhandled Exception]:", err);
      return new Response("500 Internal Server Error: Failed to process script request.\n", {
        status: 500,
        headers: { "Content-Type": "text/plain; charset=utf-8" }
      });
    }
  }
};
