<div class="feature-card rounded-xl p-6">
  <div class="flex items-start gap-4">
    <div class="w-12 h-12 rounded-lg bg-blue-500/10 flex items-center justify-center flex-shrink-0">
      <span class="text-2xl">🔍</span>
    </div>
    <div>
      <h3 class="text-xl font-semibold text-white mb-2">Semantic Search Engine</h3>
      <p class="text-slate-400 leading-relaxed">Search across all your memories with natural language queries using TF-IDF scoring, exact phrase boosting, and word-overlap fallback. The search index is cached for fast subsequent lookups.</p>
    </div>
  </div>
</div>

<div class="feature-card rounded-xl p-6">
  <div class="flex items-start gap-4">
    <div class="w-12 h-12 rounded-lg bg-emerald-500/10 flex items-center justify-center flex-shrink-0">
      <span class="text-2xl">✏️</span>
    </div>
    <div>
      <h3 class="text-xl font-semibold text-white mb-2">Edit & Delete Support</h3>
      <p class="text-slate-400 leading-relaxed">Edit existing memories and delete outdated ones directly from the CLI. No more manual file editing — <code class="text-sm bg-white/5 px-1.5 py-0.5 rounded">edit</code> and <code class="text-sm bg-white/5 px-1.5 py-0.5 rounded">delete</code> commands work non-interactively, making them agent-friendly.</p>
    </div>
  </div>
</div>

<div class="feature-card rounded-xl p-6">
  <div class="flex items-start gap-4">
    <div class="w-12 h-12 rounded-lg bg-purple-500/10 flex items-center justify-center flex-shrink-0">
      <span class="text-2xl">🤖</span>
    </div>
    <div>
      <h3 class="text-xl font-semibold text-white mb-2">Agent-First Design</h3>
      <p class="text-slate-400 leading-relaxed">Every command supports JSON output and requires no interactive prompts. Integrates naturally with Claude Code, OpenCode, and Copilot via bridge commands that auto-generate configuration files.</p>
    </div>
  </div>
</div>

<div class="feature-card rounded-xl p-6">
  <div class="flex items-start gap-4">
    <div class="w-12 h-12 rounded-lg bg-amber-500/10 flex items-center justify-center flex-shrink-0">
      <span class="text-2xl">🐛</span>
    </div>
    <div>
      <h3 class="text-xl font-semibold text-white mb-2">Bug Fixes</h3>
      <ul class="text-slate-400 leading-relaxed space-y-1 list-disc list-inside">
        <li>TF-IDF now correctly counts document frequency (was counting total occurrences, causing negative scores)</li>
        <li><code class="text-sm bg-white/5 px-1.5 py-0.5 rounded">--json</code> flag no longer leaks into search queries</li>
        <li>Search index is now cached on recall, not just on remember</li>
        <li>Word-overlap fallback handles queries like "UI design" matching "UI/Design"</li>
      </ul>
    </div>
  </div>
</div>
