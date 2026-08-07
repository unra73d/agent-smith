/**
 * Auto-load a project-local `.env` into the OpenCode process at startup.
 *
 * Why: the desktop app has no shell to `source .env`, so MCP servers configured
 * with `{env:...}` in opencode.json would otherwise receive empty values. This
 * lets opencode.json stay committable (`{env:...}`, no hardcoded secrets) while
 * the gitignored `.env` supplies the values locally.
 *
 * In CI there is no `.env` file, so this is a no-op and the real environment
 * (GitHub Actions secrets) is used. Existing environment variables are never
 * overridden, so CI/shell values always win over `.env`.
 *
 * No external dependencies — only Node built-ins (provided by the Bun runtime).
 */
import { existsSync, readFileSync } from "node:fs"
import { join } from "node:path"

function parseDotenv(text: string): Record<string, string> {
  const vars: Record<string, string> = {}
  for (const raw of text.split(/\r?\n/)) {
    const line = raw.trim()
    if (!line || line.startsWith("#")) continue
    const eq = line.indexOf("=")
    if (eq < 1) continue // need a non-empty key before '='
    const key = line.slice(0, eq).trim()
    let val = line.slice(eq + 1).trim() // keep any '=' inside the value
    if (
      val.length >= 2 &&
      ((val.startsWith('"') && val.endsWith('"')) ||
        (val.startsWith("'") && val.endsWith("'")))
    ) {
      val = val.slice(1, -1)
    }
    vars[key] = val
  }
  return vars
}

export const LoadEnv = async ({
  directory,
  worktree,
}: {
  directory?: string
  worktree?: string
}) => {
  const root = directory || worktree || process.cwd()
  const envPath = join(root, ".env")
  const vars = existsSync(envPath)
    ? parseDotenv(readFileSync(envPath, "utf8"))
    : {}

  // 1) Set into process.env at startup so opencode.json `{env:...}`
  //    interpolation can resolve the values. Never override an existing value.
  for (const [k, v] of Object.entries(vars)) {
    if (!process.env[k]) process.env[k] = v
  }

  return {
    // 2) Also inject into shell / subprocess execution (belt-and-suspenders for
    //    MCP "local" server spawns that read env at exec time).
    "shell.env": async (
      _input: unknown,
      output: { env: Record<string, string> },
    ) => {
      for (const [k, v] of Object.entries(vars)) {
        if (!output.env[k]) output.env[k] = v
      }
    },
  }
}
