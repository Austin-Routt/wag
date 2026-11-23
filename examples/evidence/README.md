# Synthetic evidence-record example

This example is synthetic. The report, participants, consents, target argument
and input pins do not represent observations or decisions in a real game.

From `wag/`:

```sh
go run ./cmd/wag-evidence ../examples/evidence/bundle.json
```

The expected result is a valid record contract with `reading-use` admitted
**for consideration** by two distinct model participants. It is not a finding
that the room was at the reported temperature. The assessment leaves source
authenticity, instrument conditions and factual checking unresolved.

The second source is an unattempted capture. It has no invented payload or
claim. No network request is made. The example demonstrates that honest
unavailability can be retained without becoming a negative evidential finding.

The text selector uses Unicode code points: the degree sign occupies one
selector position even though its UTF-8 representation uses two bytes.
Git attributes preserve the payload's exact line endings on checkouts that
otherwise convert text to CRLF. Changing the bytes legitimately changes the
payload identity and requires a new record rather than silently fixing its hash.

Changes to the offered interpretation, target, representation or context must
not inherit the existing consents. Native tests exercise those adverse cases,
withdrawal and malformed records. The full contract and remaining native-game
integration are in [the contract](../../docs/evidence-contract.md) and
[implementation plan](../../docs/evidence-implementation-plan.md).
