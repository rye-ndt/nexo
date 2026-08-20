-- The role library a new install opens with. Ids are fixed on purpose:
-- seed/workflow.sql points its steps at them. One statement, not five, so the
-- guard still reads an empty table on the last row as it did on the first: a
-- library that already holds anything is left alone. json_object builds the two
-- JSON columns so nothing here has to be escaped twice.

INSERT OR IGNORE INTO roles (
    id, name, description, effort, retryable, pause_for_review,
    inputs, instructions, output_structure, created_at, updated_at
)
SELECT * FROM (
    SELECT
    '01988000-0001-7000-8000-000000000001',
    'Explorer',
    'Reads a project and reports how it is put together.',
    'quick',
    1,
    0,
    '{}',
    json_object(
        'base',
        'You explore a codebase and explain it to someone who has never opened it. Read before you conclude, and name the files you read. Change nothing on disk.'
    ),
    'purpose: what this project is for, in one sentence
layout: the folders that matter and what lives in each
build: how it is built, tested and run
entry_points: where a reader should start
unknowns: what you could not work out from the code',
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now'),
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
    UNION ALL
    SELECT
    '01988000-0002-7000-8000-000000000002',
    'Planner',
    'Turns a request into a short ordered plan.',
    'standard',
    1,
    0,
    '{}',
    json_object(
        'base',
        'You turn a request into a plan the next step can follow. If an earlier step handed you its findings, start from those. Read enough of the project to know what is already there, then write the smallest ordered set of steps that finishes the job, and say what you are leaving out on purpose. Change nothing on disk.'
    ),
    'goal: what this plan is meant to achieve, in one sentence
steps: the work in order, one line each, small enough to check
files: the files each step is expected to touch
risks: what could go wrong and what to do about it
out_of_scope: what you decided not to do',
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now'),
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
    UNION ALL
    SELECT
    '01988000-0003-7000-8000-000000000003',
    'Implementer',
    'Makes one described change to the code.',
    'deep',
    1,
    0,
    json_object(
        'change',
        json_object(
            'description', 'What to change',
            'required', json('true'),
            'type', 'textarea',
            'default', '',
            'options', json_array()
        )
    ),
    json_object(
        'base',
        'You write code. Make this change and nothing more: {{change}}. Read the code around it first and follow the style that is already there instead of bringing your own. Build the project when you are done and fix what you broke before you report.'
    ),
    'summary: what you changed, in one sentence
files: every file you touched and what changed in it
approach: why you did it this way rather than another way
verification: the build or test commands you ran and what they printed
follow_ups: what you left for a later step',
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now'),
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
    UNION ALL
    SELECT
    '01988000-0004-7000-8000-000000000004',
    'Reviewer',
    'Reads what an earlier step did and says whether it is sound.',
    'deep',
    1,
    1,
    '{}',
    json_object(
        'base',
        'You review work another step has already done. Read the code as it now stands and report only defects you can point at a line for, worst first. Say plainly whether this should go ahead. Change nothing on disk.'
    ),
    'verdict: ship, fix first, or needs a rewrite
defects: the file and line of each one, and what makes it wrong rather than unusual
risks: what this change could break that its author may not have seen
next_steps: what the following step should do about it',
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now'),
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
    UNION ALL
    SELECT
    '01988000-0005-7000-8000-000000000005',
    'Tester',
    'Writes the tests a change is missing and runs them.',
    'standard',
    1,
    0,
    '{}',
    json_object(
        'base',
        'You write tests. Cover the behaviour a reader would doubt rather than the lines that are easy to reach, and put each test beside the code it covers. Run the suite and report failures word for word. Fix the tests you wrote, not the code under test.'
    ),
    'added: the test files and cases you wrote
command: how the suite is run
outcome: what passed and what failed
failures: each failure with the output it printed
gaps: what is still untested, and why',
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now'),
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
)
WHERE NOT EXISTS (SELECT 1 FROM roles);
