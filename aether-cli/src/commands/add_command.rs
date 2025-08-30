use std::path::PathBuf;
use crate::model::project_models::FileTracking;

pub fn execute(file: PathBuf) -> anyhow::Result<()> {

    let tracking = FileTracking::init()?;

    tracking.track(&file)?;

    println!("Tracking file {}", file.display());

    Ok(())
}