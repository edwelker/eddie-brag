# Gamification Features

Goal: Make brag documentation habitual through positive reinforcement and visible progress.

## A. Streak Tracking

- Track consecutive days/weeks with at least one entry
- Display in `brag list` header: "7-day streak"
- Warn when streak at risk: "No entries this week - streak ends in 2 days"
- Implementation: Add `last_entry_date` tracking, calculate days since

## B. Achievement Badges

`brag achievements` command. Unlockable:
- First Entry
- First 100% Complete Entry
- 10 Enriched Entries
- Process Master (5 Process entries with metrics)
- Leadership Evidence (3 Leadership entries with peer recognition)
- Week Warrior (7 consecutive days)
- Impact Tracker ($100K+ total calculated impact)
- Speed Demon (enriched entry within 24 hours of add)

Implementation:
- Add `achievements` array to BragDocument
- Check/unlock after each add/enrich
- Display unlocked at end of operations

## C. Visual Progress Dashboard (`brag stats`)

```
Brag Stats - Month 1

Total Entries: 9
Completeness:
  100%  (0 entries)
   60%+ (2 entries)
  <60%  (7 entries)

Current Streak: 3 days
Longest Streak: 5 days
Impact Measured: $351K (from hours saved x team size)
Achievements: 3/10 unlocked
```

## D. Self-Comparison (`brag compare`)

```
Month 1:  9 entries | 45% avg completeness | 351 hrs saved
Month 2: 12 entries | 72% avg completeness | 200 hrs saved
Month 3:  7 entries | 85% avg completeness | 150 hrs saved
```

## E. Nudges

- `brag list` with no entries this week: "Tip: No brags this week. Capture something today!"
- Completeness < 50%: "5 entries below 60%. Run `brag review` to improve."
- After add without enrichment: "Reminder: Enrich within 24 hours for Speed Demon achievement"

## Implementation Priority

1. Streak tracking (high motivation, low code)
2. Completeness dashboard (extend existing stats)
3. Achievement system
4. Nudges (low-friction behavior shaping)
5. Self-comparison leaderboard (advanced analytics)
