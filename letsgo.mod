// letsgo.mod

// The default matrix is five targets; windows/arm64 is the sixth that the
// GoReleaser config built, kept so the published asset list does not shrink.
build (
	linux/amd64
	linux/arm64
	darwin/amd64
	darwin/arm64
	windows/amd64
	windows/arm64
)

brew danielriddell21/tap

// Releases were marked as pre-releases by the shared workflow after the fact;
// letsgo does it as part of publishing, so promotion is still a manual step.
release prerelease=true
