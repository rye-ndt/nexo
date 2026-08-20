-- The workflow a new install opens on: a three-step read-only tour of whatever
-- project folder the user points it at. It seeds only when the user has no
-- workflows at all, so an existing install never finds a stranger in its rail.
-- Step ids and roleIds are fixed; the roleIds come from seed/roles.sql.
-- The doc column is the frontend Workflow type, verbatim, the same JSON the app
-- writes back on every edit.

INSERT INTO workflow_drafts (id, doc, updated_at)
SELECT
    '01988001-0000-7000-8000-000000000001',
    json_set(
        '{
            "id": "01988001-0000-7000-8000-000000000001",
            "name": "Start here",
            "createdAt": "",
            "railRank": 0,
            "locked": false,
            "started": false,
            "cancelled": false,
            "paused": false,
            "projectDir": "",
            "steps": [
                {
                    "id": "01988002-0001-7000-8000-000000000001",
                    "title": "See how this project is put together",
                    "prompt": "Read this project and describe how it is organised: what it does, the folders that matter, how it is built, and how it is run. Name the files you read. Change nothing.",
                    "state": "idle",
                    "position": {"x": 0, "y": 0},
                    "dependsOn": [],
                    "roleId": "01988000-0001-7000-8000-000000000001",
                    "values": {}
                },
                {
                    "id": "01988002-0002-7000-8000-000000000002",
                    "title": "Pick the changes worth making",
                    "prompt": "Start from what the previous step reported and pick the handful of changes that would help this project most. For each one, say what it improves and which files it would touch. Propose them only and change nothing on disk.",
                    "state": "idle",
                    "position": {"x": 340, "y": 0},
                    "dependsOn": ["01988002-0001-7000-8000-000000000001"],
                    "roleId": "01988000-0002-7000-8000-000000000002",
                    "values": {}
                },
                {
                    "id": "01988002-0003-7000-8000-000000000003",
                    "title": "Check the plan against the code",
                    "prompt": "Read the plan the previous step handed you, then open the files it names and check each change against what the code actually does. Flag anything already handled, anything that would break something else, and anything resting on a wrong assumption. Change nothing on disk.",
                    "state": "idle",
                    "position": {"x": 680, "y": 0},
                    "dependsOn": ["01988002-0002-7000-8000-000000000002"],
                    "roleId": "01988000-0004-7000-8000-000000000004",
                    "values": {}
                }
            ]
        }',
        '$.createdAt',
        strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
    ),
    strftime('%Y-%m-%dT%H:%M:%fZ', 'now')
WHERE NOT EXISTS (SELECT 1 FROM workflow_drafts);
