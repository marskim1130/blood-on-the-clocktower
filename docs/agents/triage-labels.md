# Triage Labels

The triage skill uses the following canonical labels:

| Role | Label | Description |
|------|-------|-------------|
| Needs triage | `needs-triage` | Maintainer needs to evaluate |
| Needs info | `needs-info` | Waiting on reporter |
| Ready for agent | `ready-for-agent` | Fully specified, AFK-ready |
| Ready for human | `ready-for-human` | Needs human implementation |
| Won't fix | `wontfix` | Will not be actioned |

## Setup

Create these labels in your GitHub repository:

```bash
gh label create needs-triage --color "D93F0B" --description "Maintainer needs to evaluate"
gh label create needs-info --color "0075CA" --description "Waiting on reporter"
gh label create ready-for-agent --color "0E8A16" --description "Fully specified, AFK-ready"
gh label create ready-for-human --color "FBCA04" --description "Needs human implementation"
gh label create wontfix --color "FFFFFF" --description "Will not be actioned"
```
