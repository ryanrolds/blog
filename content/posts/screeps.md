---
title: Screeps
published: 3019-01-22T00:16:19Z
intro: Reviewing lessons learned over 1 year of playing Screeps
---
I have not posted in a long time. It's been a crazy couple of years. We bought a house right before the pandemic, sister-in-law moved in for a while, been learning how to repair things around the house, and changed employeers. One of the things that has helped me keep my sanity is [Screeps](https://screeps.com/), an MMO for programmers.

Screeps allows programmers to write an AI that plays a game similar to a very large persistent RTS (think Starcraft, but a grid of maps and  2000 players). Players write logic that drives units, builds bases, defends against attacks (NPC & player), and raids bases (NPC and player). The major languges (JavaScript, TypeScript, Rust, Kotlin, and Python) have start kits. Any language that can be compiled to WASM is technically supported.

The best introduction to the game is the [tutorial](https://screeps.com/a/#!/sim/tutorial/1), which does not require purchasing the game. It's very basic and more an onramp to the concepts then an example of how to write your AI. As you complete the tutorial and play the game you will be checking out the [game docs](https://docs.screeps.com/index.html) and [API docs](https://docs.screeps.com/api/) very often. The docs and API is very well done.

## Getting started

The tutorial will introduce you to writting "creep" (the units in the game) logic, upgrade your Room Control Level (allows you to build additional kinds of structures), automatically spawning screeps, and defending your rooms. I stronly recommend doing the tutorial and writing an an AI that can get a room to RCL 4 and attack invaders. Once you have that initial AI claim a room in Shard 3.

The game has 4 shards (grids of rooms connected by portals). The starting player shard is Shard 3, which limits everyone's CPU time to 20ms, which allows a decently optimized AI to claim at least 5 rooms. At the time of this writting, I have 7 rooms on Shard 3 and 3 rooms on Shard 2 (not capped at 20ms and has an specially aggressive AI - Tiggabot - that attacks players AIs within ~10 rooms).

The number of rooms a player can claim is determined by their Global Control Level (GCL); GCL1 allows 1 room and GLC7 allows 7 rooms. Energy put into room controllers, which upgrades the Room Control Level, also upgrades the GCL.

Most players work on getting their first room to at least RCL 5, which is when the Terminal structure is unlocked and allows players to transfer resources instantly between rooms with a terminal. The terminal also allows an AI to place buy and sell orders for the several dozen resources in the game. Atr that point players turn towards expanding into multiple rooms if they have GCL2 or higher.

Rooms not claimed by the player can have their energy mined and taken to a nearby claimed room, this process is called "remote mining". This is also an early game strategy to aquire more engery and accelerate upgrading RCL & GCL.

## Programming

What I most enjoy about screeps is the





