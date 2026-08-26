export namespace core_itf {
	
	export class ApprovalOption {
	    id: string;
	    label: string;
	    description: string;
	    recommended: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ApprovalOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.description = source["description"];
	        this.recommended = source["recommended"];
	    }
	}
	export class DraftGraphStep {
	    id: string;
	    title: string;
	    role_id: string;
	    depends_on: string[];
	
	    static createFrom(source: any = {}) {
	        return new DraftGraphStep(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.role_id = source["role_id"];
	        this.depends_on = source["depends_on"];
	    }
	}
	export class DraftRequest {
	    name: string;
	    description: string;
	    workflow_name: string;
	    project_dir: string;
	    steps: DraftGraphStep[];
	
	    static createFrom(source: any = {}) {
	        return new DraftRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.description = source["description"];
	        this.workflow_name = source["workflow_name"];
	        this.project_dir = source["project_dir"];
	        this.steps = this.convertValues(source["steps"], DraftGraphStep);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Handoff {
	    step: string;
	    tldr: string;
	    outcome: string;
	    blockers: Record<string, string>;
	    approved_decisions: Record<string, string>;
	    rejected_decisions: Record<string, string>;
	    current_behaviors: Record<string, string>;
	    changed_behaviors: Record<string, string>;
	    must_avoid: Record<string, string>;
	    nuances: Record<string, string>;
	    known_gaps: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new Handoff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.step = source["step"];
	        this.tldr = source["tldr"];
	        this.outcome = source["outcome"];
	        this.blockers = source["blockers"];
	        this.approved_decisions = source["approved_decisions"];
	        this.rejected_decisions = source["rejected_decisions"];
	        this.current_behaviors = source["current_behaviors"];
	        this.changed_behaviors = source["changed_behaviors"];
	        this.must_avoid = source["must_avoid"];
	        this.nuances = source["nuances"];
	        this.known_gaps = source["known_gaps"];
	    }
	}

}

export namespace input_itf {
	
	export class AgentStatus {
	    name: string;
	    installed: boolean;
	    instance_count: number;
	    logged_in: boolean;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.installed = source["installed"];
	        this.instance_count = source["instance_count"];
	        this.logged_in = source["logged_in"];
	        this.version = source["version"];
	    }
	}
	export class ContextUsage {
	    total: number;
	    used: number;
	    billed: number;
	    input: number;
	    cached: number;
	
	    static createFrom(source: any = {}) {
	        return new ContextUsage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total = source["total"];
	        this.used = source["used"];
	        this.billed = source["billed"];
	        this.input = source["input"];
	        this.cached = source["cached"];
	    }
	}

}

export namespace output_itf {
	
	export class AgentDefaultInfo {
	    effort: string;
	    model: string;
	    model_label: string;
	    thinking_level: string;
	
	    static createFrom(source: any = {}) {
	        return new AgentDefaultInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.effort = source["effort"];
	        this.model = source["model"];
	        this.model_label = source["model_label"];
	        this.thinking_level = source["thinking_level"];
	    }
	}
	export class ModelOptionInfo {
	    model: string;
	    label: string;
	    harness: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelOptionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.label = source["label"];
	        this.harness = source["harness"];
	    }
	}
	export class AgentDefaultOptionsInfo {
	    efforts: string[];
	    models: ModelOptionInfo[];
	    thinking_levels: string[];
	
	    static createFrom(source: any = {}) {
	        return new AgentDefaultOptionsInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.efforts = source["efforts"];
	        this.models = this.convertValues(source["models"], ModelOptionInfo);
	        this.thinking_levels = source["thinking_levels"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class AgentInfo {
	    id: string;
	    status?: input_itf.AgentStatus;
	
	    static createFrom(source: any = {}) {
	        return new AgentInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.status = this.convertValues(source["status"], input_itf.AgentStatus);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ApprovalInfo {
	    id: string;
	    agent_id: string;
	    step_id: string;
	    kind: string;
	    question: string;
	    detail: string;
	    options: core_itf.ApprovalOption[];
	    multi_select: boolean;
	    requested_at: string;
	
	    static createFrom(source: any = {}) {
	        return new ApprovalInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.agent_id = source["agent_id"];
	        this.step_id = source["step_id"];
	        this.kind = source["kind"];
	        this.question = source["question"];
	        this.detail = source["detail"];
	        this.options = this.convertValues(source["options"], core_itf.ApprovalOption);
	        this.multi_select = source["multi_select"];
	        this.requested_at = source["requested_at"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class FileChangeInfo {
	    path: string;
	    old_path: string;
	    change_type: string;
	    additions: number;
	    deletions: number;
	    unified_diff: string;
	
	    static createFrom(source: any = {}) {
	        return new FileChangeInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.old_path = source["old_path"];
	        this.change_type = source["change_type"];
	        this.additions = source["additions"];
	        this.deletions = source["deletions"];
	        this.unified_diff = source["unified_diff"];
	    }
	}
	export class MCPServerInfo {
	    name: string;
	    url: string;
	    authorized: boolean;
	    authorized_at: string;
	    account: string;
	    kind: string;
	
	    static createFrom(source: any = {}) {
	        return new MCPServerInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.url = source["url"];
	        this.authorized = source["authorized"];
	        this.authorized_at = source["authorized_at"];
	        this.account = source["account"];
	        this.kind = source["kind"];
	    }
	}
	
	export class ModelPriceInfo {
	    model: string;
	    model_label: string;
	    input_price: string;
	    cached_input_price: string;
	    output_price: string;
	
	    static createFrom(source: any = {}) {
	        return new ModelPriceInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.model = source["model"];
	        this.model_label = source["model_label"];
	        this.input_price = source["input_price"];
	        this.cached_input_price = source["cached_input_price"];
	        this.output_price = source["output_price"];
	    }
	}
	export class RoleInputInfo {
	    description: string;
	    required: boolean;
	    type: string;
	    default: string;
	    options: string[];
	
	    static createFrom(source: any = {}) {
	        return new RoleInputInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.description = source["description"];
	        this.required = source["required"];
	        this.type = source["type"];
	        this.default = source["default"];
	        this.options = source["options"];
	    }
	}
	export class RoleInfo {
	    id: string;
	    name: string;
	    description: string;
	    effort: string;
	    retryable: boolean;
	    pause_for_review: boolean;
	    inputs: Record<string, RoleInputInfo>;
	    instructions: Record<string, string>;
	    output_structure: string;
	
	    static createFrom(source: any = {}) {
	        return new RoleInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.effort = source["effort"];
	        this.retryable = source["retryable"];
	        this.pause_for_review = source["pause_for_review"];
	        this.inputs = this.convertValues(source["inputs"], RoleInputInfo, true);
	        this.instructions = source["instructions"];
	        this.output_structure = source["output_structure"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class RunStepSpec {
	    client_id: string;
	    name: string;
	    prompt: string;
	    effort: string;
	    instructions: string[];
	    output_structure: string;
	    depends_on: string[];
	    auto_retry: boolean;
	    pause_for_review: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RunStepSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.client_id = source["client_id"];
	        this.name = source["name"];
	        this.prompt = source["prompt"];
	        this.effort = source["effort"];
	        this.instructions = source["instructions"];
	        this.output_structure = source["output_structure"];
	        this.depends_on = source["depends_on"];
	        this.auto_retry = source["auto_retry"];
	        this.pause_for_review = source["pause_for_review"];
	    }
	}
	export class RunWorkflowResult {
	    workflow_id: string;
	    step_ids: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new RunWorkflowResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workflow_id = source["workflow_id"];
	        this.step_ids = source["step_ids"];
	    }
	}
	export class RunWorkflowSpec {
	    project_dir_path: string;
	    steps: RunStepSpec[];
	
	    static createFrom(source: any = {}) {
	        return new RunWorkflowSpec(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_dir_path = source["project_dir_path"];
	        this.steps = this.convertValues(source["steps"], RunStepSpec);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class StepActivityInfo {
	    seq: number;
	    at: string;
	    text: string;
	
	    static createFrom(source: any = {}) {
	        return new StepActivityInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seq = source["seq"];
	        this.at = source["at"];
	        this.text = source["text"];
	    }
	}
	export class WorkflowDraftInfo {
	    id: string;
	    doc: string;
	    updated_at: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowDraftInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.doc = source["doc"];
	        this.updated_at = source["updated_at"];
	    }
	}
	export class WorkflowStepInfo {
	    step_id: string;
	    agent_id: string;
	    effort: string;
	    model: string;
	    model_label: string;
	    status: string;
	    handoffs: core_itf.Handoff[];
	    context_usage?: input_itf.ContextUsage;
	    spent?: input_itf.ContextUsage;
	    cost_usd: number;
	    priced: boolean;
	    activity: StepActivityInfo[];
	
	    static createFrom(source: any = {}) {
	        return new WorkflowStepInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.step_id = source["step_id"];
	        this.agent_id = source["agent_id"];
	        this.effort = source["effort"];
	        this.model = source["model"];
	        this.model_label = source["model_label"];
	        this.status = source["status"];
	        this.handoffs = this.convertValues(source["handoffs"], core_itf.Handoff);
	        this.context_usage = this.convertValues(source["context_usage"], input_itf.ContextUsage);
	        this.spent = this.convertValues(source["spent"], input_itf.ContextUsage);
	        this.cost_usd = source["cost_usd"];
	        this.priced = source["priced"];
	        this.activity = this.convertValues(source["activity"], StepActivityInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WorkflowStatusInfo {
	    workflow_id: string;
	    status: string;
	    steps: WorkflowStepInfo[];
	    tokens_billed: number;
	    tokens_input: number;
	    tokens_cached: number;
	    cost_usd: number;
	    priced: boolean;
	    started_at: string;
	    completed_at: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkflowStatusInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.workflow_id = source["workflow_id"];
	        this.status = source["status"];
	        this.steps = this.convertValues(source["steps"], WorkflowStepInfo);
	        this.tokens_billed = source["tokens_billed"];
	        this.tokens_input = source["tokens_input"];
	        this.tokens_cached = source["tokens_cached"];
	        this.cost_usd = source["cost_usd"];
	        this.priced = source["priced"];
	        this.started_at = source["started_at"];
	        this.completed_at = source["completed_at"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

