use std::{env, fs};
use std::time::SystemTime;
use anyhow::anyhow;
use crate::constants;
use crate::model::project_models::ProjectMetadata;

pub fn execute() -> anyhow::Result<()> {

    let current_dir = env::current_dir()?;
    let aether_dir = &current_dir.join(constants::AETHER_DIR);

    if aether_dir.exists() {
        return Err(anyhow!("Aether project already initialized in {}", current_dir.display()))
    }

    fs::create_dir_all(aether_dir)?;

    let metadata = ProjectMetadata {
        creation_timestamp: SystemTime::now()
    };

    fs::write(aether_dir.join("metadata.json"), serde_json::to_vec(&metadata)?)?;

    println!("Initialized aether project in {}", current_dir.display());
    Ok(())
}