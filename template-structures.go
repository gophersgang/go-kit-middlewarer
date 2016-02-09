package main

import (
	"fmt"
	"strings"
)

type TemplateCommon struct {
	BasePackage         string
	BasePackageName     string
	EndpointPackage     string
	EndpointPackageName string
	EndpointPrefix      string
	InterfaceName       string
	InterfaceNameLcase  string
}

type TemplateParam struct {
	PublicName string
	Name       string
	Type       string
}

func createTemplateParam(p Param) TemplateParam {
	return TemplateParam{
		Type: p.typ,
	}
}

type TemplateMethod struct {
	TemplateCommon
	LocalName              string
	MethodName             string
	MethodNameLcase        string
	MethodArguments        string
	MethodResults          string
	MethodResultNamesStr   string
	MethodArgumentNamesStr string
	MethodArgumentNames    []string
	MethodResultNames      []string
	Params                 []TemplateParam
	Results                []TemplateParam
}

func publicVariableName(str string) string {
	firstLetter := string([]byte{str[0]})
	rest := ""
	if len(str) > 1 {
		rest = str[1:]
	}

	return strings.ToUpper(firstLetter) + rest
}

func privateVariableName(str string) string {
	firstLetter := string([]byte{str[0]})
	rest := ""
	if len(str) > 1 {
		rest = str[1:]
	}

	return strings.ToLower(firstLetter) + rest
}

func createTemplateMethods(basePackage, endpointPackage Import, interf Interface, methods []Method, reseveredNames []string) []TemplateMethod {
	results := make([]TemplateMethod, 0, len(methods))
	for _, meth := range methods {
		var names []string
		names = append(names, reseveredNames...)
		names = append(names, meth.usedNames()...)
		var params []TemplateParam
		var methodsResults []TemplateParam

		var paramNames []string
		for _, p := range meth.params {
			paramNames = append(paramNames, p.names...)
			for _, n := range p.names {
				params = append(params, TemplateParam{
					PublicName: publicVariableName(n),
					Name:       n,
					Type:       p.typ,
				})
			}
		}

		var resultNames []string
		for _, r := range meth.results {
			resultNames = append(resultNames, r.names...)
			for _, n := range r.names {
				methodsResults = append(methodsResults, TemplateParam{
					PublicName: publicVariableName(n),
					Name:       n,
					Type:       r.typ,
				})
			}
		}

		lcaseName := determineLocalName(strings.ToLower(interf.name), reseveredNames)
		results = append(results, TemplateMethod{
			TemplateCommon: TemplateCommon{
				BasePackage:         basePackage.path,
				BasePackageName:     basePackage.name,
				EndpointPackage:     endpointPackage.path,
				EndpointPackageName: endpointPackage.name,
				EndpointPrefix:      fmt.Sprintf("/%s", strings.ToLower(interf.name)),
				InterfaceName:       interf.name,
				InterfaceNameLcase:  privateVariableName(interf.name),
			},
			MethodName:             meth.name,
			MethodNameLcase:        privateVariableName(meth.name),
			LocalName:              lcaseName,
			MethodArguments:        meth.methodArguments(),
			MethodResults:          meth.methodResults(),
			MethodArgumentNamesStr: meth.methodArgumentNames(),
			MethodResultNamesStr:   meth.methodResultNames(),
			MethodArgumentNames:    paramNames,
			MethodResultNames:      resultNames,
			Params:                 params,
			Results:                methodsResults,
		})
	}
	return results
}

type TemplateBase struct {
	TemplateCommon
	Imports            []string
	ImportsWithoutTime []string
	Methods            []TemplateMethod
}

func createTemplateBase(basePackage, endpointPackage Import, i Interface, imps []Import) TemplateBase {
	imps = filteredImports(i, imps)

	names := make([]string, 0, len(imps))
	for _, i := range imps {
		names = append(names, i.name)
	}

	var impspecs []string
	var impspecsWithoutTime []string
	for _, i := range imps {
		impspecs = append(impspecs, i.ImportSpec())
		if i.path != "time" {
			impspecsWithoutTime = append(impspecsWithoutTime, i.ImportSpec())
		}
	}

	return TemplateBase{
		TemplateCommon: TemplateCommon{
			BasePackage:         basePackage.path,
			BasePackageName:     basePackage.name,
			EndpointPackage:     endpointPackage.path,
			EndpointPackageName: endpointPackage.name,
			EndpointPrefix:      fmt.Sprintf("/%s", strings.ToLower(i.name)),
			InterfaceName:       i.name,
			InterfaceNameLcase:  privateVariableName(i.name),
		},
		Imports:            impspecs,
		ImportsWithoutTime: impspecsWithoutTime,
		Methods:            createTemplateMethods(basePackage, endpointPackage, i, i.methods, names),
	}
}

func filteredImports(i Interface, imps []Import) []Import {
	res := make([]Import, 0, len(imps))
	tmp := make([]string, 0, len(imps))
	for _, imp := range imps {
		for _, meth := range i.methods {
			for _, param := range meth.params {
				if strings.HasPrefix(param.typ, fmt.Sprintf("%s.", imp.name)) {
					if !sliceContains(tmp, imp.ImportSpec()) {
						res = append(res, imp)
						tmp = append(tmp, imp.ImportSpec())
					}
				}
			}

			for _, result := range meth.results {
				if strings.HasPrefix(result.typ, fmt.Sprintf("%s.", imp.name)) {
					if !sliceContains(tmp, imp.ImportSpec()) {
						res = append(res, imp)
						tmp = append(tmp, imp.ImportSpec())
					}
				}
			}
		}
	}
	return res
}