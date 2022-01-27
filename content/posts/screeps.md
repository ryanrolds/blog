---
title: Screeps after playing 1 year
published: 2022-01-26T22:16:19Z
intro: It's been a crazy couple of years. One of the things that have helped me keep my sanity is Screeps, an MMO for programmers.
---
I have not posted in a long time. It's been a crazy couple of years. We bought a house right before the pandemic, sister-in-law moved in for a while, learned to take care of a house, and changed employers. One of the things that have helped me keep my sanity in these interesting time is [Screeps](https://screeps.com/), an MMO for programmers.

Screeps asks programmers to create a bot that plays a massive persistent RTS (think Starcraft/Warcraft, but with a grid of maps and over 2000 players). Players write logic that drives units, builds bases, defends against attacks, and raids NPCs/players. The major languages (JavaScript, TypeScript, Rust, Kotlin, and Python) have starter kits. Any language that can be compiled to WASM is technically supported. If you've ever been playing an RTS and wished that you could write a bot that would play the game for you, Screeps is for you.

The best introduction to the game is the [tutorial](https://screeps.com/a/#!/sim/tutorial/1), which does not require purchasing the game. It's an onramp to the concepts rather than an example of how to write your bot. As you complete the tutorial, you will be frequently referencing the [game docs](https://docs.screeps.com/index.html) and [API docs](https://docs.screeps.com/api/). The documentation is very well done, and there is a [community managed wiki](https://wiki.screepspl.us/index.php/Getting_Started) with some of the meta.

## Getting started

The tutorial introduces you to writing logic for "creeps" (the units in the game), upgrading your room, automatically spawning creeps, and defending your room. Do the tutorial, write a bot to get a room to RCL 4, and defend against NPC invaders. Once you have that initial bot, claim a room in Shard 3 and don't look back.

The game has 4 shards (grids of rooms connected by portals). The starting player shard is Shard 3, limiting everyone's CPU time to 20ms. This allows a decently optimized bot to claim at least 5 rooms. At the time of writing, I have 8 rooms on Shard 3 and 4 rooms on Shard 2 (not capped at 20ms and has an incredibly aggressive bot - Tiggabot - that attacks bots within ~10 rooms).

The number of rooms a player can claim is determined by their Global Control Level (GCL); GCL1 allows 1 room, and GLC7 allows 7 rooms. The energy put into room controllers also upgrades the GCL. It's common to get a room to RCL 5, which unlocks the Terminal structure, allowing transferring resources between rooms and placing buy and sell. At that point, players turn towards expanding into multiple rooms and reacting resources.

Rooms not claimed by the player can have their energy mined and hauled to a nearby claimed room. This process is called "remote mining" and allows collecting more energy, accelerating the growth of RCL & GCL.

## My bot structure

Over the last year, I've evolved [my bot](https://github.com/ryanrolds/screeps) from doing DFS processing of behavior organized in a tree to the scheduling of "processes" that share tasks & data via priorities queues and event streams. The game is single-threaded, so the notion of "processes" and IPC feels like overkill. However, as the bot grew to dozens of procedures that were dependent on the state of other procedures, a complex and tightly coupled dependency graph emerged. By cutting direct data access to other procedures and instead of sharing data/state updates via topics/queues, the scope of changes shrank, refactoring became easier, and I broke the bot less. If you're seeing parallels between monoliths and microservices, that's intentional.

As the bot grew, another problem emerged, CPU profiling and general debugging. I reached for the enterprise standard solutions and implemented a simple tracer that supported spans, logging, and profiling. At the start of each tick, some global variables (🤨) are consumed, and the tracer is configured. The tracer is then provided to the bot's tick handler. As scheduled processes are run, new sub-spans are created, logs are written, and blocks of code are timed. Metrics are stored in a document that can be accessed by external services and written to a TSDB. Logs and reports on CPU time consumed by spans may be written to the bot's console as desired. Various options are provided to filter the logs and report output.

## Programming challenges

What I enjoyed most about Screeps is the variety of problems that have to be solved. It also encouraged me to think like a PM while still addressing toil and technical debt. If we are not enjoying what we are doing, why continue to do it?

### Behavior Trees

Early in development, I decided to use [Behavior Trees](https://www.gamedeveloper.com/programming/behavior-trees-for-ai-how-they-work) instead of Finite-State Machines (FSM) for driving creeps. Like all software projects, the early decisions focused on the biggest bang for the time required. The decision worked well, and I rarely think about switching to another strategy, like FSM or Goal Oriented Action Planning (GOAP). The majority of my creeps follow a simple loop: get the next task in a queue, move to a position, pick up a resource, move to another position, perform some action with that resource, and repeat.

The creep behaviors are very linear, so an FSM and GOAP simply weren't needed. That is not to say that when implementing advanced attack/defense logic, I won't reach for GOAP; I have not had the need for anything more complex or less ridged than Behavior Trees.

### Path Finding & Cost Matrices

Playing the game requires learning about [pathfinding](https://en.wikipedia.org/wiki/Pathfinding) and cost matrices. Creeps have to move, which requires calculating a path between their current position and a destination. When calculating these paths use case-specific policies have to be factored in: What is the maximum distance allowed? How much time can I spend calculating a path? Is a partial path allowed? Are there rooms that should not be entered? Do destructible walls block the path?

Thankfully Screeps provides a [sophisticated pathfinding API](https://docs.screeps.com/api/#PathFinder). The most important concept to understand when using this API is that most problems are solved with [cost matrices](https://docs.screeps.com/api/#PathFinder-CostMatrix). When calculating a path, the provided API will fetch a cost matrix for each room that the path crosses. The default cost matrix includes values for the terrain (plains, swamps, indestructible walls). A callback can be provided that allows the usage of custom cost matrices. It's also possible to tell the pathfinding to ignore rooms by returning `false` instead of the matrix. The ability to filter rooms as well as mark areas of rooms as blocked or high cost allows complex behavior to be implemented.

Two examples:

* Defenders - When defending a base, a cost matrix for the base footprint is calculated and temporarily cached. Calculating the footprint requires creating a cost matrix that has all walls and ramparts set to blocking. Applying a [flood fill algorithm](https://en.wikipedia.org/wiki/Flood_fill) at the base's origin, the footprint of the base is calculated. All positions outside of the base footprint are set to a higher cost (overwriting the wall/rampart cost that blocks movement). This results in defenders pooling inside of the base's walls closest to the enemy.
* Roads - It's beneficial to build roads so that resources can be hauled at maximum speed. This requires calculating a path from the Storage structure and the source of the resources. A direct path may cause the road to be close to high traffic areas (other resources, controllers being upgraded, etc...) causing traffic jams. These areas can be given a higher cost making the path go around the high traffic areas instead of through them.

Implementing different matrix transforms to solve very specific problems, like base placement (distance transform) and wall placement (max-flow min-cut), has been an enjoyable diversion from the day-to-day development of web services.

## Final Comments

If you enjoy programming and want a challenge outside of your day job, check this game out. Don't worry about players being hostile; Shard 3 is pretty chill. If you do get wiped, you still have your code and GCL. It's easy to claim another starting room and try again. Also, there are newbie and restart areas that are walled off from from the established players for up to 2 weeks.

If you have questions, join [Screep's Official Discord](https://discord.com/invite/screeps). The community is helpful and friendly.
