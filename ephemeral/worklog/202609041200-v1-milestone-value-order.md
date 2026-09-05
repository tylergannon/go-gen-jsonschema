# V1 milestone value ordering

decision: Rank milestone work by direct user adoption value; treat polish and release theater as lower priority unless it removes a concrete adoption risk.
correction: User asked to question issues added from eagerness to make v1 look polished rather than assuming every milestone issue belongs in the release gate.
decision: Keep only direct correctness, a bounded installable Codex plugin, representative consumer proof, and release/legal closeout in 1.0. Move machine protocol, full preview, fluent API, and model-backed evaluation infrastructure after 1.0.
decision: Extract the confirmed silent declaration-option omission from the fluent API proposal into focused blocker #80; correctness is valuable independently of a new declaration syntax.
review: Independent issue/body and branch-state review agreed #77 and trimmed #66 are critical, #60/#61/#73/#65 should move out, and recommended retaining a bounded #64 because agent distribution is part of the product intent.
friction: GitHub milestones expose no useful supported arbitrary issue-order API -> record the value order explicitly in the milestone description and reinforce it with v1-p0/v1-p1/v1-p2 labels.
correction: User confirmed the fluent API has enough practical authoring value for 1.0 and wants it completed this week after functionality fixes; restore #73, make Declare canonical, deprecate legacy syntax, and remove legacy forms from primary documentation.
