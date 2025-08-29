mod commands;
mod model;
mod constants;
mod errors;

use std::path::PathBuf;
use std::process::exit;
use anyhow::anyhow;
use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(name = "aether")]
#[command(about = "Aether - a distributed blender render manager")]
struct Cli {
    #[command(subcommand)]
    command: Commands,
}

#[derive(Subcommand)]
enum Commands {
    /// Initializes an aether project in current working directory
    Init,
    /// Manage aether remotes
    Remote {
        #[command(subcommand)]
        action: RemoteCommands
    },
    /// Adds a blend file to tracked files
    Add {
        /// A .blend file
        file: PathBuf,
    },
    /// Removes a blend file from tracked files
    Remove {
        /// An existing .blend file
        file: PathBuf
    },
    /// Commits and distributes all added aether files to the nodes
    Commit {
        #[clap(short, long)]
        out_file: Option<String>
    }
}

#[derive(Subcommand)]
enum RemoteCommands {
    /// Adds an aether remote
    Add {
        /// The host address of the aether remote
        node_address: String
    },
    /// Removes an aether remote
    Remove {
        node_name: String
    },
    /// Lists all existing aether remotes
    List
}
fn main() -> anyhow::Result<()> {
    let cli = Cli::parse();

    match cli.command {
        Commands::Init => commands::init_command::execute(),
        Commands::Add { file } => commands::add_command::execute(file),
        _ => fallback_command()
    }
}

// TODO: Remove when all commands are implemented
fn fallback_command() -> anyhow::Result<()> {
    Err(anyhow!("Not implemented"))
}