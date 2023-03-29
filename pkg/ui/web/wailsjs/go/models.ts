export namespace manifest {
	
	export class Context {
	    protect?: string[];
	    // Go type: Globals
	    globals?: any;
	
	    static createFrom(source: any = {}) {
	        return new Context(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.protect = source["protect"];
	        this.globals = this.convertValues(source["globals"], null);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice) {
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
	export class NetworkConfig {
	    subdomain: string;
	    pluralDns: boolean;
	
	    static createFrom(source: any = {}) {
	        return new NetworkConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.subdomain = source["subdomain"];
	        this.pluralDns = source["pluralDns"];
	    }
	}

}

export namespace ui {
	
	export class Application {
	    key: string;
	    label: string;
	    isDependency: boolean;
	    dependencyOf: {[key: string]: any};
	    data: {[key: string]: any};
	
	    static createFrom(source: any = {}) {
	        return new Application(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.label = source["label"];
	        this.isDependency = source["isDependency"];
	        this.dependencyOf = source["dependencyOf"];
	        this.data = source["data"];
	    }
	}

}

