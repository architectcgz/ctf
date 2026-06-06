#!/usr/bin/env node
import { existsSync, mkdirSync, writeFileSync } from "node:fs";
import { join, resolve } from "node:path";

const VALID_STATUSES = new Set(["not-impl", "implemented", "agent-recorded", "rejected", "archived"]);

function usage() {
  console.error(`Usage:
  node record-improvement.mjs --title "Short title" [--status not-impl] [--root .] [--body "Details"]

Statuses:
  not-impl | implemented | agent-recorded | rejected | archived`);
}

function parseArgs(argv) {
  const args = { status: "not-impl", root: ".", body: "" };
  for (let i = 0; i < argv.length; i += 1) {
    const key = argv[i];
    const value = argv[i + 1];
    if (!key.startsWith("--") || value === undefined) {
      usage();
      process.exit(2);
    }
    i += 1;
    if (key === "--title") args.title = value;
    else if (key === "--status") args.status = value;
    else if (key === "--root") args.root = value;
    else if (key === "--body") args.body = value;
    else {
      console.error(`Unknown option: ${key}`);
      usage();
      process.exit(2);
    }
  }
  return args;
}

function slugify(value) {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9\u4e00-\u9fa5]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 80) || "improvement";
}

function today() {
  return new Date().toISOString().slice(0, 10);
}

const args = parseArgs(process.argv.slice(2));
if (!args.title) {
  console.error("Missing required --title.");
  usage();
  process.exit(2);
}
if (!VALID_STATUSES.has(args.status)) {
  console.error(`Invalid --status: ${args.status}`);
  usage();
  process.exit(2);
}

const root = resolve(args.root);
const feedbackDir = join(root, "feedback");
mkdirSync(feedbackDir, { recursive: true });

const date = today();
const slug = slugify(args.title);
let filePath = join(feedbackDir, `${date}-${slug}.md`);
let suffix = 2;
while (existsSync(filePath)) {
  filePath = join(feedbackDir, `${date}-${slug}-${suffix}.md`);
  suffix += 1;
}

const content = `# ${args.title}\n\n## 问题描述\n\n${args.body || "TODO: describe what was observed."}\n\n## 原因分析\n\nTODO: explain why this matters.\n\n## 解决方案\n\nTODO: describe what should be improved.\n\n## 收获\n\n- TODO: summarize the reusable lesson.\n\n## 沉淀状态\n\n- 状态：${args.status}\n- Owner：\n- 链接：\n\n## 证据\n\n- file:\n- command:\n- behavior:\n\n## Decision Log\n\n- ${date}: Created.\n`;

writeFileSync(filePath, content);
console.log(filePath);
