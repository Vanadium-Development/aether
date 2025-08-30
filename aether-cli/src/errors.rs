use std::env::current_dir;
use anyhow::anyhow;

pub fn uninitialized_error() -> anyhow::Result<anyhow::Error> {
    Ok(anyhow!("No aether project is initialized at {}", current_dir()?.display()))
}