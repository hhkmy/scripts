/**
 * Cloudflare Worker: scripts.hhk.my.id
 * Reverse proxy for https://github.com/hhkmy/scripts
 * 
 * Features:
 * - Direct plain-text bash script serving (no redirects required)
 * - Intelligent User-Agent routing:
 *   - Terminal (curl/wget): Displays clean CLI banner & script menu
 *   - Web Browsers: Redirects to GitHub repository
 * - Cloudflare Edge Caching (5 minutes)
 * - Safe 404 error handling
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
    const url = new URL(request.url);
    const path = url.pathname.replace(/^\/+/, ""); // strip leading slash
    const userAgent = (request.headers.get("user-agent") || "").toLowerCase();
    const isCli = userAgent.includes("curl") || userAgent.includes("wget") || userAgent.includes("httpie");

    // 1. Handle Root (/)
    if (!path) {
      if (isCli) {
        return new Response(generateCliBanner(), {
          headers: {
            "Content-Type": "text/plain; charset=utf-8",
            "Cache-Control": "public, max-age=3600",
          }
        });
      }
      return Response.redirect(`https://github.com/${GITHUB_REPO}`, 302);
    }

    // 2. Ignore non-script metadata requests
    if (path === "favicon.ico" || path === "robots.txt") {
      return new Response("", { status: 404 });
    }

    // 3. Prevent directory traversal attacks
    if (path.includes("..") || path.startsWith(".") || path.includes("/")) {
      return new Response("Invalid script path\n", {
        status: 400,
        headers: { "Content-Type": "text/plain; charset=utf-8" }
      });
    }

    // 4. Fetch upstream script from raw.githubusercontent.com
    const upstreamUrl = `${UPSTREAM_BASE}/${path}`;
    const cacheKey = new Request(url.toString(), request);
    const cache = caches.default;

    // Check Cloudflare edge cache first
    let cachedResponse = await cache.match(cacheKey);
    if (cachedResponse) {
      return cachedResponse;
    }

    const upstreamResponse = await fetch(upstreamUrl, {
      headers: {
        "User-Agent": "scripts.hhk.my.id Cloudflare-Worker"
      }
    });

    if (upstreamResponse.status === 404) {
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
      return new Response(`Upstream GitHub error: ${upstreamResponse.status} ${upstreamResponse.statusText}\n`, {
        status: 502,
        headers: { "Content-Type": "text/plain; charset=utf-8" }
      });
    }

    // 5. Construct fast response with plain-text content
    const body = await upstreamResponse.text();
    const response = new Response(body, {
      status: 200,
      headers: {
        "Content-Type": "text/plain; charset=utf-8",
        "Cache-Control": "public, max-age=300, s-maxage=300", // 5 minutes cache
        "X-Content-Type-Options": "nosniff",
        "Access-Control-Allow-Origin": "*"
      }
    });

    // Store in Cloudflare edge cache asynchronously
    ctx.waitUntil(cache.put(cacheKey, response.clone()));

    return response;
  }
};
