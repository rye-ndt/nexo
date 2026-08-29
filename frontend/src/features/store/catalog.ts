/**
 * What the store sells, which is nothing: this is the ready-made work Nexo ships
 * with, hardcoded for now. Role ids are fixed and permanent — the first five are
 * the ids seed/roles.sql writes on a fresh install, so a new user honestly sees
 * them as already theirs, and adding any role twice is a no-op the library can
 * detect without storing anything extra.
 *
 * Step ids only have to be unique inside their own template: adding one runs the
 * graph through copyWorkflow, which mints fresh ids for the workflow and every
 * step in it.
 */

import {Effort, InputType} from '@/shared/lib/enums'
import type {Role} from '@/features/roles/types'
import type {StoreTemplate} from '@/features/store/types'

export const StoreRoleId = {
    Explorer: '01988000-0001-7000-8000-000000000001',
    Planner: '01988000-0002-7000-8000-000000000002',
    Implementer: '01988000-0003-7000-8000-000000000003',
    Reviewer: '01988000-0004-7000-8000-000000000004',
    Tester: '01988000-0005-7000-8000-000000000005',
    Documenter: '01988000-0006-7000-8000-000000000006',
    Debugger: '01988000-0007-7000-8000-000000000007',
    Refactorer: '01988000-0008-7000-8000-000000000008',
    Releaser: '01988000-0009-7000-8000-000000000009',
} as const

export const STORE_ROLES: Role[] = [
    {
        id: StoreRoleId.Explorer,
        name: 'Explorer',
        description: 'Reads a project and reports how it is put together.',
        effort: Effort.Quick,
        retryable: true,
        pauseForReview: false,
        inputs: [],
        instructions: [
            {
                key: 'base',
                value: 'You explore a codebase and explain it to someone who has never opened it. Read before you conclude, and name the files you read. Change nothing on disk.',
            },
        ],
        outputStructure: `purpose: what this project is for, in one sentence
layout: the folders that matter and what lives in each
build: how it is built, tested and run
entry_points: where a reader should start
unknowns: what you could not work out from the code`,
    },
    {
        id: StoreRoleId.Planner,
        name: 'Planner',
        description: 'Turns a request into a short ordered plan.',
        effort: Effort.Standard,
        retryable: true,
        pauseForReview: false,
        inputs: [],
        instructions: [
            {
                key: 'base',
                value: 'You turn a request into a plan the next step can follow. If an earlier step handed you its findings, start from those. Read enough of the project to know what is already there, then write the smallest ordered set of steps that finishes the job, and say what you are leaving out on purpose. Change nothing on disk.',
            },
        ],
        outputStructure: `goal: what this plan is meant to achieve, in one sentence
steps: the work in order, one line each, small enough to check
files: the files each step is expected to touch
risks: what could go wrong and what to do about it
out_of_scope: what you decided not to do`,
    },
    {
        id: StoreRoleId.Implementer,
        name: 'Implementer',
        description: 'Makes one described change to the code.',
        effort: Effort.Deep,
        retryable: true,
        pauseForReview: false,
        inputs: [
            {
                key: 'change',
                label: 'What to change',
                type: InputType.Textarea,
                required: true,
            },
        ],
        instructions: [
            {
                key: 'base',
                value: 'You write code. Make this change and nothing more: {{change}}. Read the code around it first and follow the style that is already there instead of bringing your own. Build the project when you are done and fix what you broke before you report.',
            },
        ],
        outputStructure: `summary: what you changed, in one sentence
files: every file you touched and what changed in it
approach: why you did it this way rather than another way
verification: the build or test commands you ran and what they printed
follow_ups: what you left for a later step`,
    },
    {
        id: StoreRoleId.Reviewer,
        name: 'Reviewer',
        description: 'Reads what an earlier step did and says whether it is sound.',
        effort: Effort.Deep,
        retryable: true,
        pauseForReview: true,
        inputs: [],
        instructions: [
            {
                key: 'base',
                value: 'You review work another step has already done. Read the code as it now stands and report only defects you can point at a line for, worst first. Say plainly whether this should go ahead. Change nothing on disk.',
            },
        ],
        outputStructure: `verdict: ship, fix first, or needs a rewrite
defects: the file and line of each one, and what makes it wrong rather than unusual
risks: what this change could break that its author may not have seen
next_steps: what the following step should do about it`,
    },
    {
        id: StoreRoleId.Tester,
        name: 'Tester',
        description: 'Writes the tests a change is missing and runs them.',
        effort: Effort.Standard,
        retryable: true,
        pauseForReview: false,
        inputs: [],
        instructions: [
            {
                key: 'base',
                value: 'You write tests. Cover the behaviour a reader would doubt rather than the lines that are easy to reach, and put each test beside the code it covers. Run the suite and report failures word for word. Fix the tests you wrote, not the code under test.',
            },
        ],
        outputStructure: `added: the test files and cases you wrote
command: how the suite is run
outcome: what passed and what failed
failures: each failure with the output it printed
gaps: what is still untested, and why`,
    },
    {
        id: StoreRoleId.Documenter,
        name: 'Documenter',
        description: 'Brings the written docs back in line with the code.',
        effort: Effort.Quick,
        retryable: true,
        pauseForReview: false,
        inputs: [
            {
                key: 'doc_path',
                label: 'The file to bring up to date',
                type: InputType.File,
                required: false,
            },
        ],
        instructions: [
            {
                key: 'base',
                value: 'You keep documentation honest. Read {{doc_path}} if it was given, otherwise find the docs this project already has. Fix every claim the code contradicts and leave the voice alone — you are correcting facts, not rewriting prose. Add nothing you cannot point at code for.',
            },
        ],
        outputStructure: `corrections: each claim you fixed, and the code that disproved the old one
added: anything the docs were missing outright
stale: what still looks wrong but you could not confirm
voice: anything you deliberately left as it was written`,
    },
    {
        id: StoreRoleId.Debugger,
        name: 'Debugger',
        description: 'Reproduces a reported fault and finds the line that causes it.',
        effort: Effort.Deep,
        retryable: true,
        pauseForReview: false,
        inputs: [
            {
                key: 'symptom',
                label: 'What goes wrong, and how to see it happen',
                type: InputType.Textarea,
                required: true,
            },
        ],
        instructions: [
            {
                key: 'base',
                value: 'You find causes, not cures. Reproduce this first: {{symptom}}. Do not propose a fix until you have seen the fault yourself. Then read back from where it surfaces to where it starts, and name the line that is actually wrong. Change nothing on disk except throwaway instrumentation you remove before you report.',
            },
        ],
        outputStructure: `reproduced: yes or no, and the exact command and output
cause: the file and line that is wrong, and why it is wrong
path: how the fault travels from that line to the symptom
fix: the smallest change that would correct it
ruled_out: what you suspected and disproved`,
    },
    {
        id: StoreRoleId.Refactorer,
        name: 'Refactorer',
        description: 'Reshapes code without changing what it does.',
        effort: Effort.Deep,
        retryable: true,
        pauseForReview: true,
        inputs: [
            {
                key: 'target',
                label: 'The file or package to reshape',
                type: InputType.Text,
                required: true,
            },
        ],
        instructions: [
            {
                key: 'base',
                value: 'You reshape {{target}} without changing what it does. Behaviour is the contract: if the tests would notice, you have gone too far. Prefer deleting to abstracting, and leave the code reading like the code around it. Run the build and the tests before and after, and report both.',
            },
        ],
        outputStructure: `shape: what the code looked like before and what it looks like now
deleted: what you removed outright
behaviour: the evidence that nothing observable changed
verification: the commands you ran before and after, with their output
left_alone: what you would have changed but did not, and why`,
    },
    {
        id: StoreRoleId.Releaser,
        name: 'Release notes',
        description: 'Turns a range of commits into notes a user can read.',
        effort: Effort.Quick,
        retryable: true,
        pauseForReview: false,
        inputs: [
            {
                key: 'since',
                label: 'The tag or commit to start from',
                type: InputType.Text,
                required: true,
                default: 'HEAD~20',
            },
        ],
        instructions: [
            {
                key: 'base',
                value: 'You write release notes for the people who use this software, not the people who wrote it. Read the commits since {{since}} and describe what changed for a user. Group by what it affects, lead with what they will notice, and drop anything purely internal. Name no commit hashes. Change nothing on disk.',
            },
        ],
        outputStructure: `headline: the one change worth leading with
added: new things a user can now do
fixed: faults that are gone, described by what used to go wrong
changed: behaviour that is different, and what to do about it
internal: what you left out on purpose`,
    },
]

export const STORE_TEMPLATES: StoreTemplate[] = [
    {
        id: '01988001-0000-7000-8000-000000000001',
        name: 'Start here',
        description:
            'Reads a project you have never run agents on, picks the changes worth making, and checks that plan against the code. Nothing is written to disk.',
        steps: [
            {
                id: 'explore',
                title: 'See how this project is put together',
                prompt: 'Read this project and describe how it is organised: what it does, the folders that matter, how it is built, and how it is run. Name the files you read. Change nothing.',
                position: {x: 0, y: 0},
                dependsOn: [],
                roleId: StoreRoleId.Explorer,
            },
            {
                id: 'plan',
                title: 'Pick the changes worth making',
                prompt: 'Start from what the previous step reported and pick the handful of changes that would help this project most. For each one, say what it improves and which files it would touch. Propose them only and change nothing on disk.',
                position: {x: 340, y: 0},
                dependsOn: ['explore'],
                roleId: StoreRoleId.Planner,
            },
            {
                id: 'check',
                title: 'Check the plan against the code',
                prompt: 'Read the plan the previous step handed you, then open the files it names and check each change against what the code actually does. Flag anything already handled, anything that would break something else, and anything resting on a wrong assumption. Change nothing on disk.',
                position: {x: 680, y: 0},
                dependsOn: ['plan'],
                roleId: StoreRoleId.Reviewer,
            },
        ],
    },
    {
        id: '01988001-0000-7000-8000-000000000002',
        name: 'Ship a small change',
        description:
            'The everyday loop: plan the change, make it, cover it with tests, then have it read back before you keep it.',
        steps: [
            {
                id: 'plan',
                title: 'Plan the change',
                prompt: 'Read enough of the project to know what is already there, then write the smallest ordered plan that delivers the change asked for. Name the files each step will touch.',
                position: {x: 0, y: 0},
                dependsOn: [],
                roleId: StoreRoleId.Planner,
            },
            {
                id: 'build',
                title: 'Make the change',
                prompt: 'Follow the plan the previous step handed you. Make that change and nothing more, in the style already in the file. Build the project and fix what you broke before you report.',
                position: {x: 340, y: 0},
                dependsOn: ['plan'],
                roleId: StoreRoleId.Implementer,
            },
            {
                id: 'cover',
                title: 'Cover it with tests',
                prompt: 'Write tests for the behaviour the previous step just added or changed. Cover what a reader would doubt, run the suite, and report failures word for word.',
                position: {x: 680, y: 0},
                dependsOn: ['build'],
                roleId: StoreRoleId.Tester,
            },
            {
                id: 'review',
                title: 'Read it back',
                prompt: 'Review the change as it now stands, including its tests. Report only defects you can point at a line for, and say plainly whether this should go ahead.',
                position: {x: 1020, y: 0},
                dependsOn: ['cover'],
                roleId: StoreRoleId.Reviewer,
            },
        ],
    },
    {
        id: '01988001-0000-7000-8000-000000000003',
        name: 'Chase down a bug',
        description:
            'Reproduce the fault before touching anything, fix the line that actually causes it, then prove the fix with a test and a second pair of eyes.',
        steps: [
            {
                id: 'find',
                title: 'Reproduce it and find the cause',
                prompt: 'Reproduce the reported fault yourself before proposing anything. Then read back from where it surfaces to where it starts and name the line that is wrong.',
                position: {x: 0, y: 120},
                dependsOn: [],
                roleId: StoreRoleId.Debugger,
            },
            {
                id: 'fix',
                title: 'Fix the cause',
                prompt: 'Make the smallest change that corrects the cause the previous step named. Do not fix the symptom somewhere else. Build the project before you report.',
                position: {x: 340, y: 120},
                dependsOn: ['find'],
                roleId: StoreRoleId.Implementer,
            },
            {
                id: 'guard',
                title: 'Write the test that would have caught it',
                prompt: 'Write the test that would have failed before the previous step and passes after it. Run the suite and report the output word for word.',
                position: {x: 680, y: 0},
                dependsOn: ['fix'],
                roleId: StoreRoleId.Tester,
            },
            {
                id: 'review',
                title: 'Read the fix back',
                prompt: 'Review the fix as it now stands. Say whether it treats the cause or the symptom, and what else it could break.',
                position: {x: 680, y: 240},
                dependsOn: ['fix'],
                roleId: StoreRoleId.Reviewer,
            },
        ],
    },
    {
        id: '01988001-0000-7000-8000-000000000004',
        name: 'Understand a codebase',
        description:
            'One read of the project, then two things done with it at once: the docs brought back in line, and a plan for where to start.',
        steps: [
            {
                id: 'explore',
                title: 'Read the project',
                prompt: 'Read this project and describe how it is organised: what it does, the folders that matter, how it is built, and how it is run. Name the files you read. Change nothing.',
                position: {x: 0, y: 120},
                dependsOn: [],
                roleId: StoreRoleId.Explorer,
            },
            {
                id: 'docs',
                title: 'Bring the docs back in line',
                prompt: 'Using what the previous step found, fix every claim the existing docs make that the code contradicts. Leave the voice alone and add nothing you cannot point at code for.',
                position: {x: 340, y: 0},
                dependsOn: ['explore'],
                roleId: StoreRoleId.Documenter,
            },
            {
                id: 'plan',
                title: 'Say where to start',
                prompt: 'Using what the first step found, write the shortest ordered path a new contributor should take through this project to become useful. Change nothing on disk.',
                position: {x: 340, y: 240},
                dependsOn: ['explore'],
                roleId: StoreRoleId.Planner,
            },
        ],
    },
    {
        id: '01988001-0000-7000-8000-000000000005',
        name: 'Clean up as you go',
        description:
            'Find the part of the project that has drifted, reshape it without changing what it does, and prove the behaviour survived.',
        steps: [
            {
                id: 'explore',
                title: 'Find what has drifted',
                prompt: 'Read this project and report the parts that have drifted furthest from the shape the rest of it keeps — duplication, dead code, a file doing three jobs. Name files and lines. Change nothing.',
                position: {x: 0, y: 0},
                dependsOn: [],
                roleId: StoreRoleId.Explorer,
            },
            {
                id: 'reshape',
                title: 'Reshape it',
                prompt: 'Take the worst of what the previous step found and reshape it without changing what it does. Prefer deleting to abstracting. Run the build and the tests before and after.',
                position: {x: 340, y: 0},
                dependsOn: ['explore'],
                roleId: StoreRoleId.Refactorer,
            },
            {
                id: 'prove',
                title: 'Prove nothing changed',
                prompt: 'Run the full suite against the reshaped code and report the output word for word. Where the previous step touched behaviour that had no test, write one.',
                position: {x: 680, y: 0},
                dependsOn: ['reshape'],
                roleId: StoreRoleId.Tester,
            },
        ],
    },
]
