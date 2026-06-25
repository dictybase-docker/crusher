// opencode plugin: wire sem CLI helpers into the agent.
// Loaded by Bun at opencode startup. Registers a thin "sem" passthrough tool
// so the agent can invoke `sem entities`, `sem diff`, `sem blame`, etc. without
// dropping into a raw bash call.
export default {
  name: "sem-tools",
  setup(opencode) {
    opencode.tool.register("sem", {
      description: "Run a sem CLI subcommand (entities, diff, blame, log, context, impact)",
      parameters: {
        type: "object",
        properties: {
          subcommand: {
            type: "string",
            enum: ["entities", "diff", "blame", "log", "context", "impact"],
            description: "sem subcommand to run",
          },
          args: {
            type: "array",
            items: { type: "string" },
            description: "Positional arguments forwarded to the subcommand",
          },
        },
        required: ["subcommand"],
      },
      async execute({ subcommand, args }) {
        const { exitCode, stdout, stderr } = await opencode.process.start({
          command: "sem",
          args: [subcommand, ...(args ?? [])],
        })
        if (exitCode !== 0) {
          return { error: stderr || `sem ${subcommand} exited with ${exitCode}` }
        }
        return { result: stdout }
      },
    })
  },
}
