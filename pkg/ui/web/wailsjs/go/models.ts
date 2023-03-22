export namespace api {
	
	export class Recipe {
	    // Go type: OIDCSettings
	    oidcSettings?: any;
	    recipeDependencies: Recipe[];
	
	    static createFrom(source: any = {}) {
	        return new Recipe(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oidcSettings = this.convertValues(source["oidcSettings"], null);
	        this.recipeDependencies = this.convertValues(source["recipeDependencies"], Recipe);
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

}

