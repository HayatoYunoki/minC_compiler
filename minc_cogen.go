package main
import (
 	"fmt"
 	//"reflect"
	"strings"
	"strconv"
)
func ast_to_asm_program(program * Program) string {
	all_asm := ""
	//前の関数が何番目の引数をreturnしたか
	return_map := 0
	free_reg_pointer := 0
	pre_return_num := 0
	pre_terms := []string{}
	//関数の個数だけ繰り返す
	for def_num, def := range program.defs {
			//asm変数に機械語プログラムを文字列として格納する
		asm := "        .text\n        .globl  "
		//get_name()より関数名を取得し、func_nameに格納
		func_name := def.get_name()
		asm += func_name
		asm += "\n        .type   "
		asm += func_name
		asm += ", @function\n"
		asm += func_name
		asm += ":\n.LFB0"+strconv.Itoa(def_num)+":\n        .cfi_startproc\n	endbr64\n"
		branch_stmts := []string{"", "", ""}
		while_branch_stmt := ""
		//is_branch := false
		//last_index_not_changed := false
		branch_num := 0
		pre_branch_num := 0
		is_rax_last := false
		new_register := 0
		register_map := map[string]string{}
		pre_stmt_type := ""
		last_asm := ""
		terms := []string{}
		push_num := 0
		//再帰に入った回数
		recur_times := 0
		use_r13 := false
		//need_pop := false
		
		// fmt.Printf("programの中身 %s\n", def.ast_to_str_def())
		body := def.get_body()
		body_type := body.get_type_name()
		//n行目がn番目に入る
		compound_parse_stmts := []string{}
		//param_locはlong型のパラメータが入るレジスタが順番に入っているリスト
		param_loc := []string{"%rdi", "%rsi", "%rdx", "%rcx", "%r8", "%r9", "8(%rsp)", "16(%rsp)", "24(%rsp)", "32(%rsp)", "40(%rsp)", "48(%rsp)"}
		//callee_save := []string{"%rbx", "%rbp", "%r12", "%r13", "%r14", "%r15"}
		free_reg := []string{"%rbx", "%r11", "%r12", "%r13", "%r14", "%r16"}
		//1つ目の関数の引数のリストparamsを取得
		params := def.get_params()
		//関数の引数からregister_mapを作る
		make_register_map(register_map, params, param_loc)
		switch body_type{
		case "StmtCompound":
			return_value, stmt_types := body.parse_stmt()
			fmt.Println(stmt_types)
			fmt.Println(return_value)
			compound_parse_stmts = strings.Split(return_value, "\n")
			types := strings.Split(stmt_types, "\n")
			//branch_num = branch_num(types)
			pre_terms0 := ""//for文の1週前のterms[0]を保存する
			for row_i, compound_parse_stmt := range compound_parse_stmts {
				//for文のこの周のasm
				this_asm := ""
				//long xなどアセンブリを発生させない行は無視する
				if compound_parse_stmt == "" {
					continue
				}
				//termsは2(以上)項演算子で式を分割したときの項のリスト
				terms = make_terms(compound_parse_stmt, true)
				//opsは2項演算子のリスト
				ops := make_ops(compound_parse_stmt, true)
				is_num := is_num_check(compound_parse_stmt)
				fmt.Println(terms)
				fmt.Println(ops)
				
				if types[row_i] == "StmtIf" || types[row_i] == "whileStmtIf" {
					//is_branch = true
					if len(terms) == 1 {
						register := param_loc[where_in_params(terms[0], params)]
						this_asm += "	testq	"+register+", "+register+"\n"
						this_asm += "	jne	.branch"+strconv.Itoa(branch_num)+strconv.Itoa(def_num)+"\n"
						
					} else if len(terms) == 2 {
						register := param_loc[where_in_params(terms[0], params)]
						if terms[1] != "0" {
							this_asm += "	cmpq	$"+terms[1]+", "+register+"\n"
							if ops[0] == "<" {
								this_asm += "	jl	.branch"+strconv.Itoa(branch_num)+strconv.Itoa(def_num)+"\n"
							}
						} else {
							this_asm += "	testq	"+register+", "+register+"\n"
						
						
							if ops[0] == "=" {
								this_asm += "	jne	.branch"+strconv.Itoa(branch_num)+strconv.Itoa(def_num)+"\n"
							} else if ops[0] == ">" {
								this_asm += "	jg	.branch"+strconv.Itoa(branch_num)+strconv.Itoa(def_num)+"\n"
							} else if ops[0] == "<" {
								this_asm += "	jle	.branch"+strconv.Itoa(branch_num)+strconv.Itoa(def_num)+"\n"
							} else if ops[0] == "<=" {
								this_asm += "	jle .branch"+strconv.Itoa(branch_num)+strconv.Itoa(def_num)+"\n"
							}
						}
					}
					branch_num += 1
					//asm += this_asm
				}
				if types[row_i] == "callfunStmtReturn" {
					this_asm += "	call	"+terms[0]+"\n"
					last_asm += "	ret\n"
				}
				if types[row_i] == "callfunStmtExpr" || types[row_i] == "elsecallfunStmtExpr"{
					if types[row_i] == "elsecallfunStmtExpr" || push_num == 0{//再起的に関数を呼び出す場合
						this_asm = "	pushq	"+free_reg[push_num]+"\n"+this_asm
						push_num += 1
						//need_pop = true
					}
					this_asm += "	call	"+terms[0]+"\n"
					this_asm += "	movq	%rax, "+free_reg[free_reg_pointer]+"\n"
					if (types[row_i] == "elsecallfunStmtExpr" || push_num == 0) && use_r13{//2つ以上の再起的に関数を呼び出す場合は引数を
						//all_asm = change_str(all_asm, 1, "$r12", "")
					} 
					// else if types[row_i] == "elsecallfunStmtExpr" || push_num == 0{
					// 	this_asm = "  pushq   "+free_reg[push_num]+"\n"+this_asm
                    //     push_num += 1
                    //     //need_pop = true
					// }
				}
				if types[row_i] == "ExprIntLiteral" || types[row_i] == "elseExprIntLiteral"{//f()+1みたいなときは1はrcxに入れると決める
					if is_num_check(terms[0]) {
						this_asm += "	movl	$"+terms[0]+", %ebx\n"
					} 
				}
				if types[row_i] == "callop" || types[row_i] == "elsecallop"{
					if ops[0] == "+" {//f()+10
						this_asm += "	addq	%rbx, %rax\n"
					} else if ops[0] == "=" {//a=f()
						//代入p=10のとき関数の引数でなければresult_locに新しい場所を用意する
						result_loc := ""
						if !is_include_map(pre_terms0, register_map) {
							result_loc = param_loc[len(params)+(new_register)]
							new_register += 1
							register_map[pre_terms0] = result_loc
						}
						result_loc = register_map[pre_terms0]
						if types[row_i] == "callop" {
							this_asm += "	movq	"+free_reg[free_reg_pointer]+", "+ result_loc+"\n"
						} else if types[row_i] == "elsecallop" {//再起的に関数を呼び出す時
							register_map[pre_terms0] = free_reg[free_reg_pointer]
						}			
						free_reg_pointer += 1
					}
				}
				if (types[row_i] == "elsecallargsStmtExpr") {//再起的に関数を呼ぶ時
					if recur_times == 1 && push_num != 0{//2つの再帰関数を呼ぶとき
						// this_asm += "	pushq	%r13\n"
						// this_asm += "	movq	%rdi, %r13\n"

						this_asm += "	subq	$16, %rsp\n"
						this_asm += "	movq	%r13, %rdi\n"
						use_r13 = true
					} else if push_num != 0{
						this_asm += "	movq	%rdi, %r13\n"
					}
					if (types[row_i+4] == "elsecallargsStmtExpr") {//次の行も再起的に関数を読む場合、引数を待避させる。
						this_asm += "  pushq   %r13\n"
                        this_asm += "    movq    %rdi, %r13\n"
					}
					this_asm += "	subq	$"+strings.Join(strings.Fields(terms[len(terms)-1]), "")+", %rdi\n"
						recur_times += 1
				}
				if (types[row_i] == "callargsStmtReturn" || types[row_i] == "callargsStmtExpr" || types[row_i] == "elsecallargsStmtExpr") && strings.Contains(terms[0], ","){//f(x, y, ...)のようなとき
					terms[0] = strings.Join(strings.Fields(terms[0]), "")//空白文字を取り除く
					func_args := strings.Split(terms[0], ",")
					for i, func_arg := range func_args {
						if i >= 6 && pre_return_num == 1 {
							if return_map >=6 && i == return_map{
								if is_num_check(func_arg) {
									//%rdiを%r10に入れる時は他より先に入れるf70
									asm += "	movq	$"+func_arg+", %r10\n"
									//どこから何個、どの文字列を何に置き換えるか
									all_asm = change_str(all_asm, 2, param_loc[return_map], "%r10")
									//param_loc[i] = callee_save[i-6]
								} else {
									asm += "	movq	"+register_map[func_arg]+", %r10\n"
									register_map[func_arg] = param_loc[return_map-6]
									all_asm = change_str(all_asm, 2, param_loc[return_map], "%r10")
								}
							}
						} else if i >= 6 {
							this_asm += "	movq	"+register_map[func_arg]+", "+free_reg[i-6]+"\n"
							all_asm = change_str(all_asm, 1, register_map[pre_terms[i]], free_reg[i-6])
							register_map[pre_terms[i]] = free_reg[i-6]
						} else {
							if is_num_check(func_arg) {
								this_asm += "	movq	$"+func_arg+", "+param_loc[i]+"\n"
							} else {
								this_asm += "	movq	"+register_map[func_arg]+", "+param_loc[i]+"\n"
							}
						}
						
					}
					fmt.Println("map")
					fmt.Println(register_map)
				} else if types[row_i] == "StmtReturn" || types[row_i] == "StmtExpr" || types[row_i] == "elseStmtExpr" || types[row_i] == "elseStmtReturn" || types[row_i] == "StmtWhile" || types[row_i] == "whileStmtExpr" || types[row_i] == "callargsStmtReturn" || types[row_i] == "whileStmtReturn" || types[row_i] == "whileStmtIf"{
					if (types[row_i] == "StmtExpr" || types[row_i] == "elseStmtExpr") && len(terms) == 2 && ops[0] == "=" {
						is_rax_last = true
					}
					//整数をreturnするとき
					if is_num == true {
						//負の整数をreturnするとき
						if compound_parse_stmt[:1] == "-" {
							this_asm += "        movq    $"
							this_asm += compound_parse_stmt
							this_asm += ", %rax\n"
							this_asm += "	ret\n"
						}else {//非負の整数をreturnするとき
							this_asm += "        movq    $"
							this_asm += compound_parse_stmt
							this_asm += ", %rax\n"
							this_asm += "	ret\n"
						}
					} else{//変数を含む値をreturnするとき
						//演算結果を格納するレジスタ名。ops[i]の結果はresult_loc[i]に入る
						result_loc := make([]string, len(ops))
						//result_locをレジスタ名として持つ変数名
						result_name := make([]string, len(ops))
						last_index := 0
						//演算子を含まない単項をreturnするとき
						if len(terms) == 1{
							this_asm += single_term_asm(compound_parse_stmt, param_loc, params)
							if compound_parse_stmt[0:1] != "-" && !is_rax_last && types[row_i] != "callargsStmtReturn"{
								this_asm += "	movq	"+param_loc[where_in_params(terms[0], params)]+", %rax\n"
							}
							//演算子が0個なので本来サイズ0であるが、例外的に
							result_loc = make([]string, 1)
							result_loc[0] = register_map[terms[0]]
							return_map = where_in_params(terms[0],params)
							
						} else  { //演算子を含む値をreturnするとき
							//比較演算子のリストcmp_list、それぞれの比較演算子がopsの何番目に入っているかを表すリストwhere_cmp
							cmp_list, where_cmp := cmp_list(ops)

							//比較演算子を含まない最初のかたまりの式を計算
							//計算するopsの範囲calc_rangeをindex2つで指定
							end := len(ops)
							if len(where_cmp) != 0 {
								end = where_cmp[0]
							}
							calc_range := []int{0, end}
							//○(%rsp)の形のレジスタ名をリネームしたかどうかを保存
							name_changed := make([]bool, len(param_loc))
							this_asm += calc_formula(param_loc, params, terms, ops, calc_range, result_loc, result_name, &last_index, name_changed, &new_register, register_map, types[row_i], def_num, branch_num)
							//比較演算子の個数回演算を行う
							for cmp_i:=0; cmp_i<len(cmp_list); cmp_i++ {
								calc_range[0] = where_cmp[cmp_i]+1
								if cmp_i == len(cmp_list)-1 {//最後の比較演算子のとき
									calc_range[1] = len(ops)
								} else {
									calc_range[1] = where_cmp[cmp_i+1]
								}
								this_asm += calc_formula(param_loc, params, terms, ops, calc_range, result_loc, result_name, &last_index, name_changed, &new_register, register_map, types[row_i], def_num, branch_num)
								cmp := cmp_list[cmp_i]
								if cmp == "<" || cmp == ">" || cmp == "<=" || cmp == ">=" || cmp == "==" || cmp == "!="{
									i := where_cmp[cmp_i]
									//iはopsの中で何番目か
									if !is_num_check(terms[i+1]) {
										result_loc[i] = param_loc[where_in_params(terms[i+1], params)]
										result_name[i] = terms[i+1]
									} else {
										result_loc[i] = param_loc[where_in_params(terms[i], params)]
										result_name[i] = terms[i]
									}
									term1 := ""//a+b*c==d+e*fのときに気をつける
									if len(ops)-len(cmp_list) >= 1 && (i-2) >=0 {
										if ops[i-1] == "*" && ops[i-2] == "+"{
											term1 = result_name[i-2]
										} else {
											term1 = result_name[i-1]
										}
									} else {
										term1 = terms[i]
									}
									this_asm += double_term_ams(term1, terms[i+1], cmp, param_loc, params, &result_loc[i], name_changed, register_map, types[row_i], def_num, branch_num)
									terms[i+1] = result_name[i]
									last_index = i
								}
								
								
							}
						}
						if (types[row_i] == "StmtReturn" || types[row_i] == "elseStmtReturn") && compound_parse_stmt[0:1] != "-" && compound_parse_stmt[0:1] != "!"{//return -xなども除外
							this_asm += "	movq	"+result_loc[last_index]+", %rax\n"
							if recur_times != 0 && push_num != 0 && use_r13 == true {
								this_asm += "	addq	$"+strconv.Itoa((push_num)*8)+", %rsp\n"
							}
							for i:=push_num-1; i>=0; i-- {
								this_asm += "	popq	"+free_reg[i]+"\n"
							}
							if recur_times != 0 && push_num != 0 && use_r13 == true{
								this_asm += "	popq %r13\n"
							}
							this_asm += "        ret\n"
						}
						if types[row_i] == "callargsStmtReturn" {
							fmt.Println("tem00")
							fmt.Println(register_map)
							this_asm += "	movq	"+register_map[terms[0]]+", %rdi\n"
						}
					}
					
					
				}
				
				fmt.Println(types[row_i])
				if pre_stmt_type == "whileStmtExpr" && types[row_i] != "whileStmtExpr" {//whileを抜けた時
					while_branch_stmt += this_asm
					asm += "	jmp		.while_in\n"
				} else if branch_num == 0 {
					asm += this_asm
				}  else {
					if branch_num != pre_branch_num {
						if pre_branch_num == 0 {
							asm += this_asm
						} else {
							branch_stmts[branch_num-2] += this_asm
						}
						pre_branch_num = branch_num
					} else {//else statementは元のブランチに
						if (branch_num - 2 < 0) && (types[row_i][:4] == "else") {
							asm += this_asm
						} else if types[row_i][:4] == "else" {
							branch_stmts[branch_num-2] += this_asm
						} else {
							branch_stmts[branch_num-1] += this_asm
						}
					}
				}
				pre_stmt_type = types[row_i]
				pre_terms0 = terms[0]
			}
			asm += last_asm
			pre_return_num = len(terms)
			pre_terms = make([]string, len(terms))
			copy(pre_terms, terms)
			fmt.Println(compound_parse_stmts)
		}
		
		for i, branch_stmt := range branch_stmts {
			asm += ".branch"+strconv.Itoa(i)+strconv.Itoa(def_num)+":\n	"+branch_stmt+"\n"
		}

		asm += ".while_branch"+strconv.Itoa(def_num)+":\n	"+while_branch_stmt+"\n"

		asm += ".not_equal"+strconv.Itoa(def_num)+":\n	movl $0, %eax\n	ret\n"
		asm += ".equal"+strconv.Itoa(def_num)+":\n	movl $1, %eax\n	ret\n"

		asm += "        .cfi_endproc\n.LFE0"+strconv.Itoa(def_num)+":\n        .size   "
		asm += func_name
		asm += ", .-"
		asm += func_name + "\n"
		
		//fmt.Printf("programの中身 %s\n", program.ast_to_str_program())
		//asm := "this is an assembly code generated by minc compiler ...\n"
		//panic("YOU MUST IMPLEMENT go/minc/minc_cogen.go:ast_to_asm_program")
		
		all_asm += asm
	}
	return all_asm
}

//minc_ast.goのDefインターフェースの定義に新しくget_name()を追加した。
//get_name()は関数名を返すメソッド
func (def * DefFun)get_name() string {
	return def.name
}

//Defインターフェースの定義に新しくget_body()を追加
//get_body()はプログラムのbodyを返すメソッド
func (def * DefFun)get_body() Stmt{
	return def.body
}

//Defインターフェースの定義に新しくget_params()を追加
//get_params()は関数の引数を返すメソッド
func (def * DefFun)get_params() []string{
	var param_names []string
	for _, param_name := range def.params {
		param_names = append(param_names, param_name.name)
	}	
	return param_names
}

//Stmtインターフェースの定義に新しくget_type_name()を追加
//get_type_name()はStmtのそれぞれの構造体について構造体名をstringで返すメソッド
func (stmt * StmtCompound)get_type_name() string{
	return "StmtCompound"
}
func (stmt * StmtEmpty)get_type_name() string{
	return "StmtEmpty"
}
func (stmt * StmtContinue)get_type_name() string{
	return "StmtContinue"
}
func (stmt * StmtBreak)get_type_name() string{
	return "StmtBreak"
}
func (stmt * StmtReturn)get_type_name() string{
	return "StmtReturn"
}
func (stmt * StmtExpr)get_type_name() string{
	return "StmtExpr"
}
func (stmt * StmtIf)get_type_name() string{
	return "StmtIf"
}
func (stmt * StmtWhile)get_type_name() string{
	return "StmtWhile"
}

//Exprインターフェースの定義に新しくget_expr_type()を追加
//get_expr_type()はExprのそれぞれの構造体について構造体名をstringで返すメソッド
func (expr * ExprIntLiteral)get_expr_type() string{
	return "ExprIntLiteral"
}
func (expr * ExprId)get_expr_type() string{
	return "ExprId"
}
func (expr * ExprOp)get_expr_type() string{
	return "ExprOp"
}
func (expr * ExprParen)get_expr_type() string{
	return "ExprParen"
}
func (expr * ExprCall)get_expr_type() string{
	return "ExprCall"
}

//Exprインターフェースの定義に新しくget_args()を追加
//get_args()はExprOpやExprCallのargsをリストで返すメソッド
func (expr * ExprIntLiteral)get_args() []Expr{
	return nil
}
func (expr * ExprId)get_args() []Expr{
	return nil
}
func (expr * ExprOp)get_args() []Expr{
	return expr.args
}
func (expr * ExprParen)get_args() []Expr{
	return nil
}
func (expr * ExprCall)get_args() []Expr{
	return nil
}

//Stmtインターフェースの定義に新しくparse_stmt()を追加
//parse_stmt()はstmtの欲しい部分だけを取り出す
func (stmt * StmtCompound)parse_stmt() (string, string){
	return_value := ""
	stmt_types := ""
	for i:=0; i<len(stmt.stmts); i++ {
		stmt_type := stmt.stmts[i].get_type_name()
		switch stmt_type{
		//return文であればStmtReturnのparse_stmt()からreturn valueを取り出す
		case "StmtReturn":
			value, this_type := stmt.stmts[i].parse_stmt()
			return_value += value+"\n"
			stmt_types += this_type+"\n"
		case "StmtExpr":
			value, this_type := stmt.stmts[i].parse_stmt()
			return_value += value+"\n"
			stmt_types += this_type+"\n"
		case "StmtIf":
			value, type_if := stmt.stmts[i].parse_stmt()
			return_value += value+"\n"
			stmt_types += type_if+"\n"
		case "StmtWhile":
			value, type_while := stmt.stmts[i].parse_stmt()
			return_value += value+"\n"
			stmt_types += type_while+"\n"
		}
	}
	return return_value, stmt_types
}

func (stmt * StmtEmpty)parse_stmt() (string, string){
	return "", "StmtEmpty"
}

func (stmt * StmtContinue)parse_stmt() (string, string){
	return "stmt", "StmtContinue"
}

func (stmt * StmtBreak)parse_stmt() (string, string){
	return "stmt", "StmtBreak"
}

func (stmt * StmtReturn)parse_stmt() (string, string){
	//「!x」を表現できるようにminc_ast.goのconcat()の1項の場合の出力を修正した
	//「-x」を表現できるようにExprOpのast_to_str_expr()を修正した
	return_value := stmt.expr.ast_to_str_expr()
	return_list := strings.Split(return_value, " ")
	//return_listが3以上のときは演算の必要がある。例：1+3など
	if len(return_list) >= 3 {
		//opに+や-などの演算子が入る
		op := return_list[1]
		//is_num[]はnum1, num2, ...がそれぞれ整数かどうかを格納する
		is_num := []bool{}
		//is_num[]に値をセットする。その際is_num_check()を使う
		for i, s := range return_list {
			if i % 2 == 0 {
				is_num = append(is_num, is_num_check(s))
			}
		}

		if all_true(is_num){
			num1, _ := strconv.Atoi(return_list[0])
			num2, _ := strconv.Atoi(return_list[2])
			//num1 + num2　をする
			if op == "+" {
				return_value = strconv.Itoa(num1 + num2)
			}
		}
	}
	//strings.Joinとstrings.Fieldsは文字列から空白を削除している
	return_value = strings.Join(strings.Fields(return_value), "")
	return_type := "StmtReturn"
	if stmt.expr.get_expr_type() == "ExprCall" {
		//関数名と引数を\nで分割
		return_value = parse_call(return_value)
		return_type = "callargs"+return_type+"\ncallfun"+return_type
	} else if stmt.expr.get_expr_type() == "ExprOp"{
		call_in := false
		op := include_op(stmt.expr.ast_to_str_expr())
		for _, arg := range stmt.expr.get_args() {
			if "ExprCall" == arg.get_expr_type() {
				//関数名と引数を\nで分割
				return_value = arg.ast_to_str_expr()
				//return_value = strings.Join(strings.Fields(return_value), "")
				return_value = parse_call(return_value)
				return_type = "callargs"+return_type+"\ncallfun"+return_type
				call_in = true
				
			} else if "ExprIntLiteral" == arg.get_expr_type() && call_in {
				return_value = return_value+"\n"+arg.ast_to_str_expr()
				//return_value += strings.Join(strings.Fields(return_value), "")
				return_type = return_type+ "\n"+"ExprIntLiteral"
				return_value += "\n"+op
				return_type += "\n"+"callop"
			}
			
		}
	} 
		
	return return_value, return_type
}

func (stmt * StmtExpr)parse_stmt() (string, string){
	//「!x」を表現できるようにminc_ast.goのconcat()の1項の場合の出力を修正した
	//「-x」を表現できるようにExprOpのast_to_str_expr()を修正した
	return_value := stmt.expr.ast_to_str_expr()
	return_list := strings.Split(return_value, " ")
	//return_listが3以上のときは演算の必要がある。例：1+3など
	if len(return_list) >= 3 {
		//opに+や-などの演算子が入る
		op := return_list[1]
		//is_num[]はnum1, num2, ...がそれぞれ整数かどうかを格納する
		is_num := []bool{}
		//is_num[]に値をセットする。その際is_num_check()を使う
		for i, s := range return_list {
			if i % 2 == 0 {
				is_num = append(is_num, is_num_check(s))
			}
		}

		if all_true(is_num){
			num1, _ := strconv.Atoi(return_list[0])
			num2, _ := strconv.Atoi(return_list[2])
			//num1 + num2　をする
			if op == "+" {
				return_value = strconv.Itoa(num1 + num2)
			}
		}
	}
	//strings.Joinとstrings.Fieldsは文字列から空白を削除している
	return_value = strings.Join(strings.Fields(return_value), "")
	return_type := "StmtExpr"
	if stmt.expr.get_expr_type() == "ExprOp"{
		call_in := false//ExprCallがあるときしか下のfor文の中身を実行しないように
		op := include_op(stmt.expr.ast_to_str_expr()) //f72では=が入る
		for i:=len(stmt.expr.get_args())-1; i>=0; i-- {
			arg := stmt.expr.get_args()[i]
			if "ExprCall" == arg.get_expr_type() {
				//関数名と引数を\nで分割
				return_value = arg.ast_to_str_expr()
				return_value = parse_call(return_value)
				return_type = "callargs"+return_type+"\ncallfun"+return_type
				call_in = true
				
			} else if "ExprId" == arg.get_expr_type() && call_in {
				return_value = return_value+"\n"+arg.ast_to_str_expr()
				return_type = return_type+ "\n"+"ExprIntLiteral"
				return_value += "\n"+op
				return_type += "\n"+"callop"
			}
			
		}
	} 

	return return_value, return_type
}

func (stmt * StmtIf)parse_stmt() (string, string){
	//「!x」を表現できるようにminc_ast.goのconcat()の1項の場合の出力を修正した
	//「-x」を表現できるようにExprOpのast_to_str_expr()を修正した
	return_value := stmt.cond.ast_to_str_expr()
	return_list := strings.Split(return_value, " ")
	//return_listが3以上のときは演算の必要がある。例：1+3など
	if len(return_list) >= 3 {
		//opに+や-などの演算子が入る
		op := return_list[1]
		//is_num[]はnum1, num2, ...がそれぞれ整数かどうかを格納する
		is_num := []bool{}
		//is_num[]に値をセットする。その際is_num_check()を使う
		for i, s := range return_list {
			if i % 2 == 0 {
				is_num = append(is_num, is_num_check(s))
			}
		}

		if all_true(is_num){
			num1, _ := strconv.Atoi(return_list[0])
			num2, _ := strconv.Atoi(return_list[2])
			//num1 + num2　をする
			if op == "+" {
				return_value = strconv.Itoa(num1 + num2)
			}
		}
	}
	//strings.Joinとstrings.Fieldsは文字列から空白を削除している
	cond := strings.Join(strings.Fields(return_value), "")
	then_stmt, then_type := stmt.then_stmt.parse_stmt()
	else_stmt := ""
	else_type := ""
	if stmt.else_stmt != nil {
		else_stmt, else_type = stmt.else_stmt.parse_stmt()
		else_types := strings.Split(else_type, "\n")
		else_type = ""
		for _, item := range else_types {
			if item != ""{
				else_type += "else" + item + "\n"
			}
		}

	}
	return_stmt := cond+"\n"+then_stmt+"\n"+else_stmt
	if return_stmt[len(return_stmt)-1:] == "\n" {//改行文字があれば取り除く
		return_stmt = return_stmt[:len(return_stmt)-1]
	}
	return_type := "StmtIf\n"+then_type+"\n"+else_type
	if return_type[len(return_type)-1:] == "\n" {//改行文字があれば取り除く
		return_type = return_type[:len(return_type)-1]
	}
	return return_stmt, return_type
}

func (stmt * StmtWhile)parse_stmt() (string, string){
	//「!x」を表現できるようにminc_ast.goのconcat()の1項の場合の出力を修正した
	//「-x」を表現できるようにExprOpのast_to_str_expr()を修正した
	return_value := stmt.cond.ast_to_str_expr()
	return_list := strings.Split(return_value, " ")
	//return_listが3以上のときは演算の必要がある。例：1+3など
	if len(return_list) >= 3 {
		//opに+や-などの演算子が入る
		op := return_list[1]
		//is_num[]はnum1, num2, ...がそれぞれ整数かどうかを格納する
		is_num := []bool{}
		//is_num[]に値をセットする。その際is_num_check()を使う
		for i, s := range return_list {
			if i % 2 == 0 {
				is_num = append(is_num, is_num_check(s))
			}
		}

		if all_true(is_num){
			num1, _ := strconv.Atoi(return_list[0])
			num2, _ := strconv.Atoi(return_list[2])
			//num1 + num2　をする
			if op == "+" {
				return_value = strconv.Itoa(num1 + num2)
			}
		}
	}
	//strings.Joinとstrings.Fieldsは文字列から空白を削除している
	cond := strings.Join(strings.Fields(return_value), "")
	body_stmt, body_type := stmt.body.parse_stmt()
	//whileのexprのtypeには頭にwhileをつける
	body_types := strings.Split(body_type, "\n")
	body_type = ""
	for _, abody := range body_types {
		if abody != "" {
			body_type += "while" + abody + "\n"
		}	
	}
	body_type = body_type[:len(body_type)-1]
	body_stmt = body_stmt[:len(body_stmt)-1]
	return cond+"\n"+body_stmt, "StmtWhile\n"+body_type
}

//strが整数であればtrue, それ以外はfalseを返す関数
func is_num_check(str string) bool {
	for _, s := range str {
		if ('0' > s || s > '9') && s != '-'{
			return false
		}
	}
	return true
}

//bool型の配列の中身が全てtrueであればtrue, それ以外はfalseを返す関数
func all_true(arr []bool) bool {
    for _, value := range arr {
        if !value {
            return false
        }
    }
    return true
}
//bool型の配列の中身が全てtrueであればtrue, それ以外はfalseを返す関数
func all_false(arr []bool) bool {
    for _, value := range arr {
        if value {
            return false
        }
    }
    return true
}

//引数に+, -, *, /, ! など含んでいる演算子を返す関数。含んでいなければ"nop"を返す
func include_op(str string) string {
	skip := 0
	for i, s := range str {
		if skip == 1 {
			skip = 0
			continue
		}
		if s == '+' {
			return "+"
		} else if s == '-' {
			return "-" 
		} else if s == '*' {
			return "*"
		} else if s == '/' {
			return "/"
		} else if s == '!' {
			if str[i+1] != '='{
				return "!"//"!"のとき
			} else {
				return "!="//"!="のとき
			}
		} else if s == '%' {
			return "%"
		} else if s == '=' {//"=="のとき
			if len(str) >= 2 {
				if str[i+1] == '='{
					skip = 1
					return "=="
				} else {
					return "="
				}
			} else {
				return "="
			}
		} else if s == '<' {
			if str[i+1] == '='{
				skip = 1
				return "<="
			} else {
				return "<"
			}
		} else if s == '>' {
			if str[i+1] == '='{
				skip = 1
				return ">="
			} else {
				return ">"
			}
		}
	}
	return "nop"
}

//リストから文字列elementを除去する
func remove_element(s []string, element string) []string {
	var result []string
	for _, v := range s {
		if v != element {
			result = append(result, v)
		}
	}
	return result
}

//termsリストを作る関数。is_firstは再起に入る前の1回目だけtrueにすることで単項を判断する。
func make_terms(formula string, is_first bool) []string {
	terms := []string{}
	sep := include_op(formula)
	sep_index := strings.Index(formula, sep)
	if sep == "nop" {
		return []string{formula}
	} else if sep == "-" && is_first == true && sep_index == 0{//-xや!xは単項とみなす
		return []string{formula}
	} else if sep == "!" && sep_index == 0 {
		return []string{formula}
	} else {
		left := make_terms(formula[:sep_index], false)
		right := []string{}
		if len(sep) == 2 {
			right = make_terms(formula[sep_index+2:], false)
		} else {
			right = make_terms(formula[sep_index+1:], false)
		}
		terms = append(left, right...)
	}
	return terms
}
//opsリストを作る関数
func make_ops(formula string, is_first bool) []string {
	ops := []string{}
	op := include_op(formula)
	op_index := strings.Index(formula, op)
	if op == "nop" {
		return []string{}
	} else if op == "-" && is_first == true && op_index == 0{//-xや!xは単項とみなす
		return []string{}
	} else if op == "!" && op_index == 0 {
		return []string{}
	} else {
		ops = append(ops, op)
		right_op := []string{}
		if len(op) == 2 {
			right_op = make_ops(formula[op_index+2:], false)
		} else {
			right_op = make_ops(formula[op_index+1:], false)
		}
		ops = append(ops, right_op...)
	}
	return ops
}

//単項termが、関数の引数の中で何番目かを返す
func where_in_params(term string, params []string) int{
	place := 0
	//単項に!や-が含まれていたら取り除く
	new_term := strings.ReplaceAll(term, "!", "") 
	new_term = strings.ReplaceAll(new_term, "-", "") 
	for i, param := range params {
		if new_term == param {
			place = i
		}
	}
	return place
}

//単項演算のアセンブリを生成する関数
func single_term_asm(term string, param_loc []string, params []string) string{
	asm := ""
	op := include_op(term)
	if op == "-" {
		asm += "	movq	%rdi, %rax\n"
		asm += "	negq	%rax\n"
		asm += "	ret\n"
	} else if op == "!" {
		term = strings.ReplaceAll(term, "!", "")
		term_place := where_in_params(term, params)
		asm += "	testq	"+param_loc[term_place]+", "+param_loc[term_place]+"\n"
		asm += "	sete	%al\n"
		asm += "	movzbq	%al, "+param_loc[term_place]+"\n"
	}
	return asm
}

//2項演算のアセンブリを生成する関数
func double_term_ams(term1 string, term2 string, op string, param_loc []string, params []string, result_loc *string, name_changed []bool, register_map map[string]string, stmt_type string, def_num int, branch_num int) string {

	asm := ""
	param_loc1 := register_map[term1]
	param_loc2 := register_map[term2]
	//!xのparam_locを求める
	if term1[0:1] == "!" {
		param_loc1 = register_map[term1[1:]]
	}
	if term2[0:1] == "!" {
		param_loc2 = register_map[term2[1:]]
	}
	//value1, value2はterm1, term2が整数であれば"$整数"が、変数であればparam_loc1, 2が入る
	value1 := "$"+term1
	value2 := "$"+term2
	if !is_num_check(term1) {
		value1 = param_loc1
	}
	if !is_num_check(term2) {
		
		value2 = param_loc2
	}

	//演算やmovの出力先レジスタが○○(%rsp)ではエラーになるので変える
	if len(value1) >= 7 {
		//以前の計算で使ったレジスタに移し替える
		param_place := get_index(value1, param_loc)
		if name_changed[param_place] == false {
			asm += "	movq	"+value1+", "+param_loc[param_place-6]+"\n"
			name_changed[param_place] = true
			register_map[term1] = param_loc[param_place-6]
		}
		*result_loc = param_loc[param_place-6]
		value1 = param_loc[param_place-6]
	}
	//以前使った出力先レジスタの変換を次の演算でも利用
	if len(value2) >= 7 {
		param_place := get_index(value2, param_loc)
		if name_changed[param_place] == false {
			asm += "	movq	"+value2+", "+param_loc[param_place-6]+"\n"
			name_changed[param_place] = true
			register_map[term2] = param_loc[param_place-6]
		}
		*result_loc = param_loc[param_place-6]
		value2 = param_loc[param_place-6]
	}

	if op == "+" {
		if value1[:1] == "$" {//10+xのようなとき
			fmt.Println("ininin")
			asm += "	addq	"+value1+", "+value2+"\n"
			*result_loc = value2
		}else {
			fmt.Println("aaa")
			fmt.Println(register_map)
			asm += "	addq	"+value2+", "+value1+"\n"
			asm += "	movq	"+value1+", "+*result_loc+"\n"
		}
	}  else if op == "-" {
		asm += "	subq	"+value2+", "+value1+"\n"
		asm += "	movq	"+value1+", "+*result_loc+"\n"
	}else if op == "*" {
		asm += single_term_asm(term2, param_loc, params)
		asm += "	imulq	"+value2+", "+value1+"\n"
		asm += "	movq	"+value1+", "+(*result_loc)+"\n"
	} else if op == "/" {
		asm += "	movq	"+value1+", %rax\n"
		asm += "	movq $0, %rdx\n"
		asm += "	movq "+value2+", %rbx\n"
		asm += "    divq %rbx\n"
		asm += "	movq %rax, "+(*result_loc)+"\n"
	}  else if op == "%" {
		asm += 	"	movq	"+value1+", %rax\n"
		asm += "	cqto\n"
		if is_num_check(term2) {
			asm += "	movl	"+value2+", %ecx\n"
			asm += "	idivq	%rcx\n"
		} else {
			asm += "	idivq	"+value2+"\n"
		}
		asm += "	movq	%rdx, "+(*result_loc)+"\n"
	} else if op == "=" {
		asm += "	movq	"+value2+", "+(*result_loc)+"\n"
		
	} else {//比較演算子の場合
		if stmt_type == "StmtWhile" {
			asm += ".while_in:\n"
		}
		asm += "	cmpq	"+value2+", "+value1+"\n"
		if stmt_type == "StmtWhile" {//whileの条件文を判定する時
			//成立していなければ飛ばす
			if op == "==" {
				asm += "	jne		.while_branch"+strconv.Itoa(def_num)+"\n"
			} else if op == "!=" {
				asm += "	je		.while_branch"+strconv.Itoa(def_num)+"\n"
			} else if op == "<" {
				asm += "	jge		.while_branch"+strconv.Itoa(def_num)+"\n"
			} else if op == ">" {
				asm += "	jle		.while_branch"+strconv.Itoa(def_num)+"\n"
			} else if op == "<=" {
				asm += "	jg		.while_branch"+strconv.Itoa(def_num)+"\n"
			} else if op == ">=" {
				asm += "	jl		.while_branch"+strconv.Itoa(def_num)+"\n"
			}	
		} else if stmt_type  == "whileStmtIf" {
			//成立していなければ飛ばす
			if op == "==" {
				asm += "	jne	.branch"+strconv.Itoa(branch_num-1)+strconv.Itoa(def_num)+"\n"
			} else if op == "!=" {
				asm += "	je		.branch"+strconv.Itoa(branch_num-1)+strconv.Itoa(def_num)+"\n"
			} else if op == "<" {
				asm += "	jge		.branch"+strconv.Itoa(branch_num-1)+strconv.Itoa(def_num)+"\n"
			} else if op == ">" {
				asm += "	jle		.branch"+strconv.Itoa(branch_num-1)+strconv.Itoa(def_num)+"\n"
			} else if op == "<=" {
				asm += "	jg		.branch"+strconv.Itoa(branch_num-1)+strconv.Itoa(def_num)+"\n"
			} else if op == ">=" {
				asm += "	jl		.branch"+strconv.Itoa(branch_num-1)+strconv.Itoa(def_num)+"\n"
			}
		} else {
			if op == "==" {
				asm += "	jne		.not_equal"+strconv.Itoa(def_num)+"\n"
			} else if op == "!=" {
				asm += "	je		.not_equal"+strconv.Itoa(def_num)+"\n"
			} else if op == "<" {
				asm += "	jge		.not_equal"+strconv.Itoa(def_num)+"\n"
			} else if op == ">" {
				asm += "	jle		.not_equal"+strconv.Itoa(def_num)+"\n"
			} else if op == "<=" {
				asm += "	jg		.not_equal"+strconv.Itoa(def_num)+"\n"
			} else if op == ">=" {
				asm += "	jl		.not_equal"+strconv.Itoa(def_num)+"\n"
			}
			asm += "	movq	$1, "+(*result_loc)+"\n"
		}
	}
	return asm
}

//比較演算子のリストcmp_listを作成し、それぞれの比較演算子がopsの何番目に入っているかを表すリストwhere_cmpも作る
func cmp_list(ops []string) ([]string, []int) {
	cmp_list := []string{}
	where_cmp := []int{}
	for i, op := range ops {
		if op == "<" || op == ">" || op == "<=" || op == ">=" || op == "==" || op == "!="{
			cmp_list = append(cmp_list, op)
			where_cmp = append(where_cmp, i)
		}
	}
	return cmp_list, where_cmp
}

//比較演算子を含まない１つの式を計算する
func calc_formula(param_loc []string, params []string, terms []string, ops []string, calc_range []int, result_loc []string, result_name []string, last_index *int, name_changed []bool, new_register *int, register_map map[string]string, stmt_type string, def_num int, branch_num int) string {
	asm := ""
	for i, op := range ops[calc_range[0]: calc_range[1]] {
		i += calc_range[0]
		if op == "*" || op == "/" || op == "%" {
			if !is_num_check(terms[i+1]) {
				result_loc[i] = register_map[terms[i]]
				result_name[i] = terms[i]
			} else {
				result_loc[i] = register_map[terms[i]]
				result_name[i] = terms[i]
			}
			asm += double_term_ams(terms[i], terms[i+1], op, param_loc, params, &result_loc[i], name_changed, register_map, stmt_type, def_num, branch_num)
			terms[i+1] = result_name[i]
			*last_index = i
		}
	}
	for i, op := range ops[calc_range[0]: calc_range[1]] {
		i += calc_range[0]
		//ops[i]の計算をするときresult_loc[i-1], result_loc[i+1]に値が入っていればそれを使う
		if i != 0 {
			if result_loc[i-1] != ""{
				terms[i-1] = result_name[i-1]
			}
		}
		if i!= len(ops)-1 {
			if result_loc[i+1] != "" {
				terms[i+1] = result_name[i+1]
			}
		}
		if op == "+" || op == "-" {
			if !is_num_check(terms[i+1]) {
				result_loc[i] = register_map[terms[i]]
				result_name[i] = terms[i]
			} else {
				result_loc[i] = register_map[terms[i]]
				result_name[i] = terms[i]
			}
			asm += double_term_ams(terms[i], terms[i+1], op, param_loc, params, &result_loc[i], name_changed, register_map, stmt_type, def_num, branch_num)
			terms[i+1] = result_name[i]
			*last_index = i
		}
	}
	for i, op := range ops[calc_range[0]: calc_range[1]] {
		if op == "=" {
			//代入p=10のとき関数の引数でなければresult_locに新しい場所を用意する
			if !is_include_map(terms[i], register_map) {
				result_loc[i] = param_loc[len(params)+(*new_register)]
				*new_register += 1
				register_map[terms[i]] = result_loc[i]
				fmt.Println(register_map)
			}
			result_loc[i] = register_map[terms[i]]
			result_name[i] = terms[i]
			term2 := terms[i+1]
			if len(ops) > i+2 {//多項式を代入する場合
				if ops[i+1] == "+" && ops[i+2] == "%"{
					term2 = terms[i+3]
				}
			}
			asm += double_term_ams(terms[i], term2, op, param_loc, params, &result_loc[i], name_changed, register_map, stmt_type, def_num, branch_num)
			terms[i+1] = result_name[i]
			*last_index = i
		}
	}
	return asm
}

func branch_num(stmt_types []string) int {
	count := 0
	for _, stmt_type := range stmt_types {
		if stmt_type == "StmtIf" {
			count += 1
		}
	}
	return count
}

//引数の文字列が、引数のリストに含まれているかどうか
func is_include_map(s string, m map[string]string) bool{
	if _, ok := m[s]; ok {
		return true
	} 
	return false
}

func is_include_list(s string, l []string) bool {
	for _, item := range l {
		if item == s {
			return true
		}
	}
	return false
}

//register_mapを作る
func make_register_map(register_map map[string]string, params []string, param_loc []string) {
	for i, param := range params {
		register_map[param] = param_loc[i]
	}
}

//関数を関数名と引数で分ける
func parse_call(call_func string) string {
	ret_str := ""
	ret_strs := strings.Split(call_func, "(")
	ret_str += ret_strs[1][:len(ret_strs[1])-1] + "\n" + ret_strs[0]
	return ret_str
}

//どこから何個、どの文字列を何に置き換えるか
func change_str(all_asm string, n int, from string, to string) string{
	count := strings.Count(all_asm, from)
	if count == 0 || n <= 0 {
		return all_asm
	}
	if n >= count {
		return strings.ReplaceAll(all_asm, from, to)
	}
	var result string
	for i := 0; i < n; i++ {
		index := strings.Index(all_asm, from)
		if index == -1 {
			break
		}
		result += all_asm[:index] + to
		all_asm = all_asm[index+len(from):]
	}

	result += all_asm
	return result
}

//リストないから要素のindexを取得
func get_index(item string, list []string) int {
	for i, s := range list {
		if item == s {
			return i
		}
	}
	return 0
}