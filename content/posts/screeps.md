---
title: Screeps after one year
published: 2022-01-26T22:16:19Z
intro: It's been a crazy couple of years. One of the things that have helped me keep my sanity is Screeps, an MMO for programmers.
---
It's been a crazy couple of years. We bought a house right before the pandemic, sister-in-law moved in for a while, learned to take care of a house, and changed employers. One of the things that have helped me keep my sanity in these interesting times is [Screeps](https://screeps.com/), an MMO for programmers.

Screeps asks programmers to create a bot that plays a massive persistent Real-time Strategy (RTS) game. Think StarCraft/Warcraft, but with a grid of maps and over 2000 players. The bot contains logic that drives units, builds bases, defends against attacks, and raids NPCs/bots. The major languages (JavaScript, TypeScript, Rust, Kotlin, and Python) have starter kits. Any language that compiles to WASM is technically supported. If you've ever been playing an RTS and wished to write a bot that would play the game, Screeps is for you.

## Getting started

The best introduction to the game is the [tutorial](https://screeps.com/a/#!/sim/tutorial/1), which doesn't require purchasing the game. It's an on-ramp to the game's concepts rather than an example of how to write your bot. As you complete the tutorial, you will frequently reference the [game docs](https://docs.screeps.com/index.html) and [API docs](https://docs.screeps.com/api/). The documentation is well done, and there is a [community-managed wiki](https://wiki.screepspl.us/index.php/Getting_Started) with some of the meta.

## Progression

The tutorial introduces you to writing logic for "creeps" (the units in the game), upgrading your room, automatically spawning creeps, and defending your room. Do the tutorial, write a bot to get a room to RCL 4, and protect against NPC invaders. Once you have that initial bot, claim a room in the starting shard, Shard 3. The shard enforces a 20ms [soft-limit](https://docs.screeps.com/cpu-limit.html) on CPU time.

The number of rooms a bot can claim is determined by their Global Control Level (GCL); GCL 1 allows 1 room, and GCL 7 allows 7 rooms. The energy put into room controllers also upgrades the GCL. Rooms not claimed by a bot can have their energy collected and hauled to a nearby claimed room. This process is called "remote mining" and helps accelerate the growth of RCL & GCL. Getting a room to RCL 6 allows the building of a Terminal structure, supports the transfer of resources between rooms and placing market orders. At that point, players start to distribute resources across rooms, react resources, and dabble in automated trading.

Once a bot can create and distribute "boosts" (materials that give significant bonuses to attacking, healing, mining, and other actions), players can focus on defense logic and sieging NPCs/bots. When the bot feels the squeeze of the 20ms CPU limit and the low-hanging optimizations have been made, it's time to expand into other shards.

## Challenges

I enjoy the variety of problems that have to be solved and learning new techniques. The game encouraged me to think like a PM while still addressing toil and technical debt. If we are not enjoying the game, why continue to do it?

### Creep Behavior

Early in development, I decided to use [Behavior Trees](https://www.gamedeveloper.com/programming/behavior-trees-for-ai-how-they-work) to drive my creeps. I considered other options: [Finite-State Machines](https://en.wikipedia.org/wiki/Finite-state_machine) (FSM) and [Goal Oriented Action Planning](https://medium.com/@vedantchaudhari/goal-oriented-action-planning-34035ed40d0b) (GOAP).

Most of my creeps follow a simple loop:

1. Get task from a queue
2. Move to a position
3. Pick up a resource
4. Move to another position
5. Perform some action with that resource
6. Go to Step 1

The majority of logic only needs to perform single actions (pick up, drop off), repeat an action (moving), and create sequences. Behavior Trees afford these basic patterns and allow them to be composed into trees. On creation, creeps are assigned a role that determines the behavior tree used to drive the creep.

Like all software projects, the early decisions focused on the biggest bang for the time. I have not needed anything more complex or less rigid than Behavior Trees. Not to say that when implementing advanced attack/defense logic, I won't reach for GOAP in the future. However, I rarely think about switching because behavior trees get the job done.

### Path Finding & Cost Matrices

Playing the game requires learning about [pathfinding](https://en.wikipedia.org/wiki/Pathfinding) and cost matrices. Creeps have to move, which requires calculating a path between their current position and a destination. Use case-specific policies must be factored during the calculation: What is the maximum distance allowed? How much time can the bot spend calculating a path? Is a partial path helpful? Which rooms should not be entered? Do destructible walls block the path? Are there areas or hostile creeps that it should avoid?

Thankfully Screeps provides a [sophisticated pathfinding API](https://docs.screeps.com/api/#PathFinder). One of the most important concepts to understand when using this API is [cost matrices](https://docs.screeps.com/api/#PathFinder-CostMatrix). When calculating a path, the bot will fetch a cost matrix for each room in the path. The default cost matrix includes values for the terrain (plains, swamps, indestructible walls). A callback can be provided that allows the usage of custom cost matrices. It's also possible to tell the pathfinding to ignore rooms by returning `false` instead of a cost matrix. The ability to filter rooms as well as mark areas of rooms as higher cost allows complex pathing to be implemented.

Two examples:

* Defenders - A cost matrix for the base footprint is calculated and temporarily cached when defending a base. Determining the footprint requires creating a cost matrix with all walls and ramparts set to blocking and applying a [flood fill algorithm](https://web.archive.org/web/20210516141251/http://www.williammalone.com/articles/html5-canvas-javascript-paint-bucket-tool/) at the base's origin. Then, a new cost matrix is created, and all positions outside the footprint are set to a higher cost. This results in defenders pooling inside the base's walls closest to the enemy.
* Roads - It's beneficial to build roads so that resources can be hauled at maximum speed. This requires calculating a path from the Storage structure and the source of resources. A direct path may cause the road to be close to high-traffic areas (other resources, controllers being upgraded, etc...), causing traffic jams. These areas can be given a higher cost making the path go around the high-traffic area instead of through it.

Implementing different matrix transforms to solve problems, like base placement (distance transform) and wall placement (max-flow min-cut), has been an enjoyable diversion from the day-to-day building of web services.

### Bot structure & Performance

Over the last year, I evolved [my bot](https://github.com/ryanrolds/screeps) from doing depth-first processing of procedures organized in a tree to the scheduling of "processes" that share tasks & data via priority queues and event streams. The game is single-threaded, so the notion of processes and IPC feels like overkill. However, as the bot grew to dozens of procedures dependent on the state of other procedures, a complex and tightly coupled dependency graph emerged. By cutting direct data access to other procedures and sharing data/state updates via topics/queues, the scope of changes shrank, refactoring became more manageable, and I deployed broken versions of the bot less. If you see parallels with monoliths and microservices, that's intentional.

Another problem emerged, CPU profiling and general debugging. I reached for the enterprise standard solutions and implemented a simple "tracer" that supports spans, logging, and profiling. At the start of each tick, some global variables (🤨) are consumed, and a new tracer is configured. The tracer is then provided to the bot's tick handler. As scheduled processes are run, sub-spans are created, logs are written, and code blocks are timed. Metrics are stored in a document that can be accessed by external services and written to a Time Series DB (TSDB). Logs and reports on CPU time consumed by spans may be written to the bot's console as desired. Various options are provided to filter the logs and report output.

## Final Comments

Check out this game if you enjoy programming and want a challenge outside of your day job. Don't worry about hostile bots; Shard 3 is pretty chill. If you do get wiped, you still have your code and GCL. It's easy to claim another starting room and try again. Also, there are newbie and restart areas walled off from the established bots for up to 2 weeks.

Join [Screeps' Official Discord](https://discord.com/invite/screeps). The community is helpful and friendly.
