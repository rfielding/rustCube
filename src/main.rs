use std::collections::HashMap;
use std::io::{self, Write};

struct Cube {
    adj: HashMap<String, Vec<String>>, // Adjacency relations
    bak: HashMap<String, String>,      // Opposite face relations
    stickers: HashMap<String, String>, // Sticker positions and their faces
}


impl Cube {
    fn new() -> Self {
        let mut adj = HashMap::new();
        let mut bak = HashMap::new();

        // Initialize adjacencies (counter-clockwise order)
        adj.insert("f".to_string(), vec!["u".to_string(), "l".to_string(), "d".to_string(), "r".to_string()]);
        adj.insert("u".to_string(), vec!["b".to_string(), "l".to_string(), "f".to_string(), "r".to_string()]);
        adj.insert("r".to_string(), vec!["u".to_string(), "f".to_string(), "d".to_string(), "b".to_string()]);
        adj.insert("b".to_string(), vec!["r".to_string(), "d".to_string(), "l".to_string(), "u".to_string()]);
        adj.insert("d".to_string(), vec!["r".to_string(), "f".to_string(), "l".to_string(), "b".to_string()]);
        adj.insert("l".to_string(), vec!["b".to_string(), "d".to_string(), "f".to_string(), "u".to_string()]);

        // Initialize opposites
        bak.insert("u".to_string(), "d".to_string());
        bak.insert("r".to_string(), "l".to_string());
        bak.insert("f".to_string(), "b".to_string());
        bak.insert("d".to_string(), "u".to_string());
        bak.insert("l".to_string(), "r".to_string());
        bak.insert("b".to_string(), "f".to_string());

        // Initialize stickers based on adjacencies
        let stickers = Self::initialize_stickers(&adj);
        assert_eq!(stickers.len(), 54);

        Cube { adj, bak, stickers }
    }

    fn initialize_stickers(adj: &HashMap<String, Vec<String>>) -> HashMap<String, String> {
        let mut stickers = HashMap::new();

        for face in adj.keys() {
            // Add center sticker
            stickers.insert(face.clone(), face.clone());

            // Add edge stickers
            if let Some(adj_faces) = adj.get(face) {
                for adj_face in adj_faces {
                    let key = format!("{}{}", face, adj_face);
                    stickers.insert(key, face.clone());
                }
            }

            // Add corner stickers
            if let Some(adj_faces) = adj.get(face) {
                for i in 0..adj_faces.len() {
                    let next = &adj_faces[(i + 1) % adj_faces.len()];
                    let key = format!("{}{}{}", face, next, adj_faces[i]);
                    stickers.insert(key, face.clone());
                }
            }
        }

        stickers
    }

    fn draw(&self, cmd: &str, repeats: i32) {
        let s = |pos: &str| self.stickers.get(pos).unwrap();
        println!("      {}{}{}      ",
            s("bul"),s("bu"),s("bru"),
        );
        println!();
        println!("      {}{}{}      ",
            s("ulb"),s("ub"),s("ubr"),
        );
        println!("      {}{}{}      ",
            s("ul"),s("u"),s("ur"),
        );
        println!("      {}{}{}      ",
            s("ufl"),s("uf"),s("urf"),
        );
        println!();
        println!("{} {}{}{} {}{}{} {}{}{} {}",
            s("bul"),
            s("lbu"),s("lu"),s("luf"),
            s("flu"),s("fu"),s("fur"),
            s("rfu"),s("ru"),s("rub"),
            s("bru"),
        );
        println!("{} {}{}{} {}{}{} {}{}{} {}",
            s("bl"),
            s("lb"),s("l"),s("lf"),
            s("fl"),s("f"),s("fr"),
            s("rf"),s("r"),s("rb"),
            s("br"),
        );
        println!("{} {}{}{} {}{}{} {}{}{} {}",
            s("bld"),
            s("ldb"),s("ld"),s("lfd"),
            s("fdl"),s("fd"),s("frd"),
            s("rdf"),s("rd"),s("rbd"),
            s("bdr"),
        );
        println!();
        println!("      {}{}{}      ",
            s("dlf"),s("df"),s("dfr"),
        );
        println!("      {}{}{}      ",
            s("dl"),s("d"),s("dr"),
        );
        println!("      {}{}{}      ",
            s("dbl"),s("db"),s("drb"),
        );
        println!();
        println!("      {}{}{}      ",
            s("bld"),s("bd"),s("bdr"),
        );
        println!();
        println!("cmd: {} x {}", cmd, repeats);
    }

    
    // 1 turn of 1 face at i
    fn turn_raw(&mut self, f: &str) {
        let adjf = self.adj.get(f).unwrap();
        for n in 0..3 {
            let i = adjf[n].clone();
            let j = adjf[(n+1)% 4].clone();
            let k = adjf[(n+3)% 4].clone();

            let e0a = format!("{}{}", f, i);
            let e1a = format!("{}{}", i, f);
            let c0a = format!("{}{}{}", f, i, k);
            let c1a = format!("{}{}{}", i, k, f);
            let c2a = format!("{}{}{}", k, f, i);

            let e0b = format!("{}{}", f, j);
            let e1b = format!("{}{}", j, f);
            let c0b = format!("{}{}{}", f, j, i);
            let c1b = format!("{}{}{}", j, i, f);
            let c2b = format!("{}{}{}", i, f, j);

            // swap a with b
            let tmp = self.stickers.get(&e0a).unwrap().clone();
            self.stickers.insert(e0a.clone(), self.stickers.get(&e0b).unwrap().clone());
            self.stickers.insert(e0b.clone(), tmp);

            let tmp = self.stickers.get(&e1a).unwrap().clone();
            self.stickers.insert(e1a.clone(), self.stickers.get(&e1b).unwrap().clone());
            self.stickers.insert(e1b.clone(), tmp);

            let tmp = self.stickers.get(&c0a).unwrap().clone();
            self.stickers.insert(c0a.clone(), self.stickers.get(&c0b).unwrap().clone());
            self.stickers.insert(c0b.clone(), tmp);

            let tmp = self.stickers.get(&c1a).unwrap().clone();
            self.stickers.insert(c1a.clone(), self.stickers.get(&c1b).unwrap().clone());
            self.stickers.insert(c1b.clone(), tmp);

            let tmp = self.stickers.get(&c2a).unwrap().clone();
            self.stickers.insert(c2a.clone(), self.stickers.get(&c2b).unwrap().clone());
            self.stickers.insert(c2b.clone(), tmp);
            
        }
    }

    // 1 turn of 1 face at i
    fn turn_raw_center(&mut self, f: &str) {
        let adjf = self.adj.get(f).unwrap();
        for n in 0..3 {
            let i = adjf[n].clone();
            let j = adjf[(n+1)% 4].clone();
            let k = adjf[(n+3)% 4].clone();

            let e0a = format!("{}{}", i, k);
            let e1a = format!("{}{}", k, i);
            let ma = format!("{}", i);

            let e0b = format!("{}{}", j, i);
            let e1b = format!("{}{}", i, j);
            let mb = format!("{}", j);

            // swap a with b 
            let tmp = self.stickers.get(&e0a).unwrap().clone();
            self.stickers.insert(e0a.clone(), self.stickers.get(&e0b).unwrap().clone());
            self.stickers.insert(e0b.clone(), tmp);

            let tmp = self.stickers.get(&e1a).unwrap().clone();
            self.stickers.insert(e1a.clone(), self.stickers.get(&e1b).unwrap().clone());
            self.stickers.insert(e1b.clone(), tmp);

            let tmp = self.stickers.get(&ma).unwrap().clone();
            self.stickers.insert(ma.clone(), self.stickers.get(&mb).unwrap().clone());
            self.stickers.insert(mb.clone(), tmp);
        }
    }

    fn turn_all(&mut self, f: &str, i: i32) {
      // turn the ENTIRE cube
      for _ in 0..i {
        self.turn_raw(f);
        self.turn_raw_center(f);
        //// because a turn of -1 is same as turn of 3
        //// and we turn opposite face
        let b = &self.bak[f].clone();
        self.turn_raw(b);
        self.turn_raw(b);
        self.turn_raw(b);
      }
    }

    fn turn(&mut self, f: &str, i: i32) {
        for _ in 0..i {
            self.turn_raw(f);
        }
    }

    fn parse_input(input: &str) -> Vec<String> {
        let mut tokens = Vec::new();
        let mut chars = input.chars().peekable();
        
        while let Some(ch) = chars.next() {
            match ch {
                'u' | 'r' | 'f' | 'd' | 'l' | 'b' | 'q' | 'n' | '/' |
                'U' | 'R' | 'F' | 'D' | 'L' | 'B' => {
                    tokens.push(ch.to_string());
                }
                '(' => {
                    tokens.push("(".to_string());
                }
                ')' => {
                    tokens.push(")".to_string());
                }
                ' ' => continue, // Ignore spaces
                _ => println!("?"),
            }
        }
        tokens
    }
    
    fn str_in_set(token: &str, set: &[&str]) -> bool {
        set.contains(&token)
    }
    
    fn parse_rewrite(tokens: Vec<String>) -> Vec<String> {
        // First do a pass to remove non-functional tokens
        let mut cleaned = Vec::new();
        for token in tokens {
            if Self::str_in_set(&token, &[
                //"(", ")",
                 "U", "D", "L", "R", "F", "B", 
                 "u", "d", "l", "r", "f", "b",
                  "/", "n", "q"]) {
                cleaned.push(token);
            }
        }
        cleaned
    }
    
    fn process_tokens(cube: &mut Cube, tokens: Vec<String>) {
        let mut turns = 1;
        let mut quit = false;
    
        let tokens2 = Self::parse_rewrite(tokens);
    
        for token in tokens2 {
            if token == "q" {
                quit = true;
                break;
            } else if token == "/" {
                turns = 3;
            } else if token == "n" {
                *cube = Self::new();
                turns = 1;
            } else if Self::str_in_set(&token, &["U", "R", "F", "D", "L", "B"]) {
                let face = token.to_lowercase();
                cube.turn_all(&face, turns);
                turns = 1;
            } else if Self::str_in_set(&token, &["u", "r", "f", "d", "l", "b"]) {
                cube.turn(&token, turns);
                turns = 1;
            }
        }
    
        if quit {
            std::process::exit(0);
        }
    }    

    fn input(&mut self) {
        let mut current = String::new();
        let mut repeats = 1;
    
        println!("Use Singmaster convention for moves: u r f d l b");
        println!("Use capital letters to turn entire cube: U R F D L B");
        println!("Use '/' for inverse of a single face (3 turns)");
        //println!("Use parens for decoration or inverse: u r /(r u) is the same as: u r /u /r");
        println!("Use 'n' for a new cube");
        println!("Use 'q' to exit");
    
        self.draw(&current, repeats);
        loop {
            print!("% ");
            io::stdout().flush().unwrap();
    
            let prev = &current.to_string();
            current.clear();
            io::stdin().read_line(&mut current).expect("Failed to read line");
    
            current = current.trim().to_string();
            
            if current.is_empty() {
               current = prev.to_string();
            }
            
            if current != prev.to_string() {
                repeats = 1;
            } else {
                repeats += 1;
            }
    
            let tokens = Self::parse_input(&current);
            Self::process_tokens(self, tokens);
    
            self.draw(&current, repeats);
        }
    }
}


fn main() {
    let mut cube = Cube::new();
    cube.input();
}