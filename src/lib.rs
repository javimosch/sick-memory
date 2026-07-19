/// sick-memory: File-based memory system for AI coding agents.
///
/// Centralized storage, git-based scoping, and worktree support.
/// All functional code is in Go (see main.go). This crate exists
/// solely for documentation tooling compatibility.

/// Crate version, kept in lock-step with the Go release.
pub const VERSION: &str = env!("CARGO_PKG_VERSION");

/// Returns the crate version string.
pub fn version() -> &'static str {
    VERSION
}

pub fn placeholder() {}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn placeholder_does_not_panic() {
        placeholder();
    }

    #[test]
    fn version_matches_package_version() {
        assert_eq!(version(), env!("CARGO_PKG_VERSION"));
    }
}
